package tars

import (
	"encoding/json"
	"math/rand"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/TarsCloud/TarsGo/tars/protocol/res/endpointf"
	"github.com/TarsCloud/TarsGo/tars/protocol/res/queryf"
	"github.com/TarsCloud/TarsGo/tars/util/consistenthash"
	"github.com/TarsCloud/TarsGo/tars/util/endpoint"
	"github.com/serialx/hashring"
)

var (
	enableSubset = true
	subsetMg     = &subsetManager{}
)

type hashString string

func (h hashString) String() string {
	return string(h)
}

type subsetConf struct {
	enable    bool
	ruleType  string // ratio/key
	ratioConf *ratioConfig
	keyConf   *keyConfig

	lastUpdate time.Time
}

type ratioConfig struct {
	ring *hashring.HashRing
}

type keyRoute struct {
	action string
	value  string
	route  string
}

type keyConfig struct {
	rules        []keyRoute
	defaultRoute string
}

type subsetManager struct {
	lock  *sync.RWMutex
	cache map[string]*subsetConf

	registry *queryf.QueryF
}

func (s *subsetManager) getSubsetConfig(servantName string) *subsetConf {
	s.lock.RLock()
	defer s.lock.RUnlock()
	var ret *subsetConf
	if v, ok := s.cache[servantName]; ok {
		ret = v
		if v.lastUpdate.Add(time.Second * 10).After(time.Now()) {
			return ret
		}
	}
	// get config from registry
	conf := &endpointf.SubsetConf{}
	retVal, err := s.registry.FindSubsetConfigById(servantName, conf)
	if err != nil || retVal != 0 {
		// log error
		return ret
	}

	ret = &subsetConf{
		ruleType:   conf.RuleType,
		lastUpdate: time.Now(),
	}
	s.cache[servantName] = ret
	// parse subset conf
	if !conf.Enable {
		ret.enable = false
		return ret
	}
	if conf.RuleType == "ratio" {
		kv := make(map[string]int)
		json.Unmarshal([]byte(conf.RuteData), &kv)
		ret.ratioConf = &ratioConfig{ring: hashring.NewWithWeights(kv)}
	} else {
		keyConf := &keyConfig{}
		kvlist := make([]map[string]string, 0)
		json.Unmarshal([]byte(conf.RuteData), &kvlist)
		for _, kv := range kvlist {
			if vv, ok := kv["default"]; ok {
				keyConf.defaultRoute = vv
			}
			if vv, ok := kv["match"]; ok {
				keyConf.rules = append(keyConf.rules, keyRoute{
					action: "match",
					value:  vv,
					route:  kv["route"],
				})
			} else if vv, ok := kv["equal"]; ok {
				keyConf.rules = append(keyConf.rules, keyRoute{
					action: "equal",
					value:  vv,
					route:  kv["route"],
				})
			}
		}
		ret.keyConf = keyConf
	}
	return ret
}

func (s *subsetManager) getSubset(servantName, routeKey string) string {
	// check subset config exists
	subsetConf := subsetMg.getSubsetConfig(servantName)
	if subsetConf == nil {
		return ""
	}
	// route key to subset
	if subsetConf.ruleType == "ratio" {
		return subsetConf.ratioConf.findSubet(routeKey)
	}
	return subsetConf.keyConf.findSubet(routeKey)
}

func subsetEndpointFilter(servantName, routeKey string, eps []endpoint.Endpoint) []endpoint.Endpoint {
	if !enableSubset {
		return eps
	}
	subset := subsetMg.getSubset(servantName, routeKey)
	if subset == "" {
		return eps
	}

	ret := make([]endpoint.Endpoint, 0)
	for i := range eps {
		if eps[i].Subset == subset {
			ret = append(ret, eps[i])
		}
	}
	return ret
}

func subsetHashEpFilter(servantName, routeKey string, m *consistenthash.ChMap) *consistenthash.ChMap {
	if !enableSubset {
		return m
	}
	subset := subsetMg.getSubset(servantName, routeKey)
	if subset == "" {
		return m
	}

	ret := consistenthash.NewChMap(32)
	for _, v := range m.GetNodes() {
		vv, ok := v.(endpoint.Endpoint)
		if ok && vv.Subset == subset {
			ret.Add(vv)
		}
	}
	return ret
}

func (k *ratioConfig) findSubet(key string) string {
	// 为空时使用随机方式
	if key == "" {
		key = strconv.Itoa(rand.Int())
	}
	v, _ := k.ring.GetNode(key)
	return v
}

func (k *keyConfig) findSubet(key string) string {
	for _, v := range k.rules {
		if v.action == "equal" && key == v.value {
			return v.route
		} else if v.action == "match" {
			if matched, _ := regexp.Match(v.value, []byte(key)); matched {
				return v.route
			}
		}
	}
	return k.defaultRoute
}
