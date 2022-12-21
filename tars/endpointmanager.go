package tars

import (
	"encoding/json"
	"fmt"
	"hash/crc32"
	"io/ioutil"
	"math/rand"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/TarsCloud/TarsGo/tars/protocol/res/endpointf"
	"github.com/TarsCloud/TarsGo/tars/protocol/res/queryf"
	"github.com/TarsCloud/TarsGo/tars/selector/consistenthash"
	"github.com/TarsCloud/TarsGo/tars/selector/modhash"
	"github.com/TarsCloud/TarsGo/tars/selector/roundrobin"
	"github.com/TarsCloud/TarsGo/tars/util/endpoint"
	"github.com/TarsCloud/TarsGo/tars/util/gtime"
)

// EndpointManager interface of naming system
type EndpointManager interface {
	SelectAdapterProxy(msg *Message) (*AdapterProxy, bool)
	GetAllEndpoint() []*endpoint.Endpoint
	preInvoke()
	postInvoke()
	addAliveEp(ep endpoint.Endpoint)
}

var (
	gManager         *globalManager
	gManagerInitOnce sync.Once
)

type globalManager struct {
	eps                 map[string]*tarsEndpointManager
	mlock               *sync.Mutex
	refreshInterval     int
	checkStatusInterval int
}

func initOnceGManager(refreshInterval int, checkStatusInterval int) {
	gManagerInitOnce.Do(func() {
		gManager = &globalManager{refreshInterval: refreshInterval, checkStatusInterval: checkStatusInterval}
		gManager.eps = make(map[string]*tarsEndpointManager)
		gManager.mlock = &sync.Mutex{}
		go gManager.updateEndpoints()
		go gManager.checkEpStatus()
	})
}

// GetManager return a endpoint manager from global endpoint manager
func GetManager(comm *Communicator, objName string, opts ...EndpointManagerOption) EndpointManager {
	// tars
	initOnceGManager(comm.Client.RefreshEndpointInterval, comm.Client.CheckStatusInterval)
	g := gManager
	g.mlock.Lock()
	key := objName + ":" + comm.hashKey()
	for _, opt := range opts {
		opt.key(&key)
	}

	if v, ok := g.eps[key]; ok {
		g.mlock.Unlock()
		return v
	}
	g.mlock.Unlock()

	TLOG.Debug("Create endpoint manager for ", objName)
	em := newTarsEndpointManager(objName, comm, opts...) // avoid dead lock
	g.mlock.Lock()
	if v, ok := g.eps[key]; ok {
		g.mlock.Unlock()
		return v
	}
	g.eps[key] = em
	err := em.doFresh()
	// if fresh is error,we should get it from cache
	if err != nil {
		for _, cache := range appCache.ObjCaches {
			if objName == cache.Name && comm.GetLocator() == cache.Locator {
				em.activeEpf = cache.Endpoints
				newEps := make([]endpoint.Endpoint, len(em.activeEpf))
				for i, ep := range em.activeEpf {
					newEps[i] = endpoint.Tars2endpoint(ep)
				}
				em.firstUpdateActiveEp(newEps)
				TLOG.Debugf("init endpoint %s %v %v", objName, em.activeEp, em.inactiveEpf)
			}
		}
	}
	g.mlock.Unlock()
	return em
}

func (g *globalManager) checkEpStatus() {
	loop := time.NewTicker(time.Duration(g.checkStatusInterval) * time.Millisecond)
	for range loop.C {
		g.mlock.Lock()
		eps := make([]*tarsEndpointManager, 0)
		for _, v := range g.eps {
			if v.locator != nil {
				eps = append(eps, v)
			}
		}
		g.mlock.Unlock()
		for _, e := range eps {
			e.checkStatus()
		}
	}
}

func (g *globalManager) updateEndpoints() {
	loop := time.NewTicker(time.Duration(g.refreshInterval) * time.Millisecond)
	for range loop.C {
		g.mlock.Lock()
		eps := make([]*tarsEndpointManager, 0)
		for _, v := range g.eps {
			if v.locator != nil {
				eps = append(eps, v)
			}
		}
		g.mlock.Unlock()
		TLOG.Debugf("start refresh %d endpoints %d", len(eps), g.refreshInterval)
		for _, e := range eps {
			err := e.doFresh()
			if err != nil {
				TLOG.Errorf("obj: %s update endpoint error: %v.", e.objName, err)
			}
		}

		// cache to file
		cfg := GetServerConfig()
		if cfg != nil && cfg.DataPath != "" {
			cachePath := filepath.Join(cfg.DataPath, cfg.Server) + ".tarsdat"
			appCache.ModifyTime = gtime.CurrDateTime
			objCache := make([]ObjCache, len(eps))
			for i, e := range eps {
				objCache[i].Name = e.objName
				objCache[i].Locator = e.comm.GetLocator()
				objCache[i].Endpoints = e.activeEpf
				objCache[i].InactiveEndpoints = e.inactiveEpf
			}
			appCache.ObjCaches = objCache
			data, _ := json.MarshalIndent(&appCache, "", "    ")
			ioutil.WriteFile(cachePath, data, 0644)
		}
	}
}

// tarsEndpointManager is a struct which contains endpoint information.
type tarsEndpointManager struct {
	objName     string // name only, no ip list
	enableSet   bool
	setDivision string
	directProxy bool
	comm        *Communicator
	locator     *queryf.QueryF

	epList      *sync.Map
	epLock      *sync.Mutex
	activeEp    []endpoint.Endpoint
	activeEpf   []endpointf.EndpointF
	inactiveEpf []endpointf.EndpointF
	rand        *rand.Rand

	checkAdapterList *sync.Map
	checkAdapter     chan *AdapterProxy

	weightType         endpoint.WeightType
	activeEpRoundRobin *roundrobin.RoundRobin
	activeEpConHash    *consistenthash.ConsistentHash
	activeEpModHash    *modhash.ModHash
	freshLock          *sync.Mutex
	lastInvoke         int64
	invokeNum          int32
}

type EndpointManagerOption interface {
	apply(e *tarsEndpointManager)
	key(k *string)
}

type OptionFunc struct {
	applyFunc func(*tarsEndpointManager)
	keyFunc   func(*string)
}

func (f OptionFunc) apply(e *tarsEndpointManager) {
	if f.applyFunc != nil {
		f.applyFunc(e)
	}
}

func (f OptionFunc) key(e *string) {
	if f.keyFunc != nil {
		f.keyFunc(e)
	}
}

func newOptionFunc(applyFunc func(*tarsEndpointManager), keyFunc func(*string)) OptionFunc {
	return OptionFunc{applyFunc: applyFunc, keyFunc: keyFunc}
}

func WithSet(setDivision string) OptionFunc {
	return newOptionFunc(func(e *tarsEndpointManager) {
		if setDivision != "" {
			e.enableSet = true
			e.setDivision = setDivision
		}
	}, func(s *string) {
		*s = *s + ":" + setDivision
	})
}

func newTarsEndpointManager(objName string, comm *Communicator, opts ...EndpointManagerOption) *tarsEndpointManager {
	if objName == "" {
		return nil
	}
	e := &tarsEndpointManager{}
	e.comm = comm
	e.freshLock = &sync.Mutex{}
	e.epList = &sync.Map{}
	e.epLock = &sync.Mutex{}
	e.checkAdapterList = &sync.Map{}
	e.rand = rand.New(rand.NewSource(time.Now().UnixNano()))
	for _, opt := range opts {
		opt.apply(e)
	}
	pos := strings.Index(objName, "@")
	if pos > 0 {
		// [direct]
		e.objName = objName[0:pos]
		endpoints := objName[pos+1:]
		e.directProxy = true
		ends := strings.Split(endpoints, ":")
		eps := make([]endpoint.Endpoint, len(ends))
		for i, end := range ends {
			eps[i] = endpoint.Parse(end)
		}
		e.firstUpdateActiveEp(eps)
	} else {
		// [proxy] TODO singleton
		TLOG.Debug("proxy mode:", objName)
		e.objName = objName
		e.directProxy = false
		obj, _ := e.comm.GetProperty("locator")
		e.locator = new(queryf.QueryF)
		TLOG.Debug("string to proxy locator ", obj)
		e.comm.StringToProxy(obj, e.locator)
		e.checkAdapter = make(chan *AdapterProxy, 1000)
	}
	return e
}

// GetAllEndpoint returns all endpoint information as a array(support not tars service).
func (e *tarsEndpointManager) GetAllEndpoint() []*endpoint.Endpoint {
	eps := e.activeEp[:]
	out := make([]*endpoint.Endpoint, len(eps))
	for i := 0; i < len(eps); i++ {
		out[i] = &eps[i]
	}
	return out
}

func (e *tarsEndpointManager) checkStatus() {
	// only in active epf need to check.
	for _, ef := range e.activeEpf {
		ep := endpoint.Tars2endpoint(ef)
		if v, ok := e.epList.Load(ep.Key); ok {
			adp := v.(*AdapterProxy)
			if e.comm.Client.KeepAliveInterval > 0 {
				adp.doKeepAlive()
			}

			firstTime, needCheck := adp.checkActive()
			if !firstTime && !needCheck {
				continue
			}

			if firstTime {
				e.epLock.Lock()
				for i := range e.activeEp {
					if e.activeEp[i] == ep {
						e.activeEp = append(e.activeEp[:i], e.activeEp[i+1:]...)
						break
					}
				}
				e.epLock.Unlock()

				e.activeEpRoundRobin.Remove(ep)
				e.activeEpConHash.Remove(ep)
				e.activeEpModHash.Remove(ep)
			}

			if needCheck {
				if _, ok := e.checkAdapterList.Load(ep.Key); !ok {
					adp := v.(*AdapterProxy)
					e.checkAdapterList.Store(ep.Key, adp)
					e.checkAdapter <- adp
					TLOG.Errorf("checkStatus|insert check adapter, ep: %+v", ep.Key)
				}
			}
		}
	}
}

func (e *tarsEndpointManager) addAliveEp(ep endpoint.Endpoint) {
	e.epLock.Lock()
	sortedEps := e.activeEp[:]
	sortedEps = append(sortedEps, ep)
	sort.Slice(sortedEps, func(i int, j int) bool {
		return crc32.ChecksumIEEE([]byte(sortedEps[i].Key)) < crc32.ChecksumIEEE([]byte(sortedEps[j].Key))
	})
	e.activeEp = sortedEps
	e.activeEpRoundRobin.Add(ep)
	e.activeEpConHash.Add(ep)
	e.activeEpModHash.Add(ep)
	e.epLock.Unlock()
}

// SelectAdapterProxy returns the selected adapter.
func (e *tarsEndpointManager) SelectAdapterProxy(msg *Message) (*AdapterProxy, bool) {
	e.epLock.Lock()
	eps := e.activeEp[:]
	e.epLock.Unlock()

	if e.directProxy && len(eps) == 0 {
		return nil, false
	}
	if !e.directProxy && len(e.activeEpf) == 0 {
		return nil, false
	}
	select {
	case adp := <-e.checkAdapter:
		TLOG.Errorf("SelectAdapterProxy|check adapter, ep: %+v", adp.GetPoint())
		e.checkAdapterList.Delete(endpoint.Tars2endpoint(*adp.GetPoint()).Key)
		return adp, true
	default:
	}
	var (
		adp *AdapterProxy
		ep  endpoint.Endpoint
		err error
	)
	if msg.isHash && msg.hashType == ConsistentHash {
		ep, err = e.activeEpConHash.Select(msg) // ConsistentHash
	} else if msg.isHash && msg.hashType == ModHash {
		ep, err = e.activeEpModHash.Select(msg) // ModHash
	} else {
		ep, err = e.activeEpRoundRobin.Select(msg) // RoundRobin
	}
	if err != nil {
		TLOG.Errorf("SelectAdapterProxy|enableWeight: %b, isHash: %b, hashType: %s, err: %+v", e.enableWeight(), msg.isHash, msg.hashType, err)
		goto random
	}
	if v, ok := e.epList.Load(ep.Key); ok {
		adp = v.(*AdapterProxy)
	} else {
		epf := endpoint.Endpoint2tars(ep)
		adp = NewAdapterProxy(e.objName, &epf, e.comm)
		e.epList.Store(ep.Key, adp)
	}
random:
	if adp == nil && !e.directProxy {
		// not any node is alive, just select a random one.
		randomEpf := e.activeEpf[e.rand.Intn(len(e.activeEpf))]
		randomEp := endpoint.Tars2endpoint(randomEpf)
		if v, ok := e.epList.Load(randomEp.Key); ok {
			adp = v.(*AdapterProxy)
		} else {
			adp = NewAdapterProxy(e.objName, &randomEpf, e.comm)
			e.epList.Store(randomEp.Key, adp)
		}
	}
	return adp, false
}

func (e *tarsEndpointManager) doFresh() error {
	if e.directProxy {
		return nil
	}
	e.freshLock.Lock()
	defer e.freshLock.Unlock()
	return e.findAndSetObj(e.locator)
}

func (e *tarsEndpointManager) preInvoke() {
	atomic.AddInt32(&e.invokeNum, 1)
	e.lastInvoke = gtime.CurrUnixTime
}

func (e *tarsEndpointManager) postInvoke() {
	atomic.AddInt32(&e.invokeNum, -1)
}

func (e *tarsEndpointManager) findAndSetObj(q *queryf.QueryF) error {
	activeEp := make([]endpointf.EndpointF, 0)
	inactiveEp := make([]endpointf.EndpointF, 0)
	var enableSet, ok bool
	var setDivision string
	var ret int32
	var err error
	if e.enableSet && e.setDivision != "" {
		enableSet = e.enableSet
		setDivision = e.setDivision
	} else if enableSet, ok = e.comm.GetPropertyBool("enableset"); ok {
		setDivision, _ = e.comm.GetProperty("setdivision")
	}

	if enableSet {
		ret, err = q.FindObjectByIdInSameSet(e.objName, setDivision, &activeEp, &inactiveEp)
	} else {
		ret, err = q.FindObjectByIdInSameGroup(e.objName, &activeEp, &inactiveEp)
	}
	if err != nil {
		TLOG.Errorf("findAndSetObj %s fail, error: %v", e.objName, err)
		return err
	}
	if ret != 0 {
		return fmt.Errorf("findAndSetObj %s fail, ret: %d", e.objName, ret)
	}

	if reflect.DeepEqual(&activeEp, &e.activeEpf) {
		TLOG.Debugf("endpoint not change: %s, set: %s", e.objName, setDivision)
		return nil
	}

	if len(activeEp) == 0 {
		TLOG.Errorf("findAndSetObj %s, empty of active endpoint", e.objName)
		return nil
	}
	TLOG.Debugf("findAndSetObj|call FindObjectById ok, obj: %s, ret: %d, active: %v, inactive: %v", e.objName, ret, activeEp, inactiveEp)

	newEps := make([]endpoint.Endpoint, len(activeEp))
	for i, ep := range activeEp {
		newEps[i] = endpoint.Tars2endpoint(ep)
	}

	// delete useless cache
	e.epList.Range(func(key, value interface{}) bool {
		flagActive := false
		flagInactive := false

		for _, ep := range newEps {
			if key == ep.Key {
				flagActive = true
				break
			}
		}
		for _, epf := range inactiveEp {
			tep := endpoint.Tars2endpoint(epf)
			if key == tep.Key {
				flagInactive = true
				break
			}
		}
		if !flagActive && !flagInactive {
			value.(*AdapterProxy).Close()
			e.epList.Delete(key)
			TLOG.Debugf("findAndSetObj|delete useless endpoint from epList: %+v", key)
		}
		return true
	})

	bSameType, lastType := true, newEps[0].WeightType
	// delete active endpoint which status is false
	sortedEps := make([]endpoint.Endpoint, 0)
	for _, ep := range newEps {
		if v, ok := e.epList.Load(ep.Key); ok {
			adp := v.(*AdapterProxy)
			if adp.status {
				sortedEps = append(sortedEps, ep)
			}
		} else {
			sortedEps = append(sortedEps, ep)
		}

		// check weightType
		if ep.WeightType != lastType {
			bSameType = false
		}
	}

	e.weightType = endpoint.ELoop
	if bSameType {
		e.weightType = endpoint.WeightType(lastType)
	}

	// make endpoint slice sorted
	sort.Slice(sortedEps, func(i int, j int) bool {
		return crc32.ChecksumIEEE([]byte(sortedEps[i].Key)) < crc32.ChecksumIEEE([]byte(sortedEps[j].Key))
	})

	roundRobinSelector := roundrobin.New(e.enableWeight())
	roundRobinSelector.Refresh(sortedEps)
	conHashSelector := consistenthash.New(e.enableWeight(), consistenthash.KetamaHash)
	conHashSelector.Refresh(sortedEps)
	modHashSelector := modhash.New(e.enableWeight())
	roundRobinSelector.Refresh(sortedEps)

	e.epLock.Lock()
	e.activeEpf = activeEp
	e.inactiveEpf = inactiveEp
	e.activeEp = sortedEps
	e.activeEpRoundRobin = roundRobinSelector
	e.activeEpConHash = conHashSelector
	e.activeEpModHash = modHashSelector
	e.epLock.Unlock()

	TLOG.Debugf("findAndSetObj|activeEp: %+v", sortedEps)
	return nil
}

func (e *tarsEndpointManager) firstUpdateActiveEp(eps []endpoint.Endpoint) {
	if len(eps) == 0 {
		return
	}
	bSameType, lastType := true, eps[0].WeightType
	sortedEps := make([]endpoint.Endpoint, 0, len(eps))
	for _, ep := range eps {
		sortedEps = append(sortedEps, ep)
		// check weightType
		if ep.WeightType != lastType {
			bSameType = false
		}
	}

	e.weightType = endpoint.ELoop
	if bSameType {
		e.weightType = endpoint.WeightType(lastType)
	}

	// make endpoint slice sorted
	sort.Slice(sortedEps, func(i int, j int) bool {
		return crc32.ChecksumIEEE([]byte(sortedEps[i].Key)) < crc32.ChecksumIEEE([]byte(sortedEps[j].Key))
	})
	roundRobinSelector := roundrobin.New(e.enableWeight())
	roundRobinSelector.Refresh(sortedEps)
	conHashSelector := consistenthash.New(e.enableWeight(), consistenthash.KetamaHash)
	conHashSelector.Refresh(sortedEps)
	modHashSelector := modhash.New(e.enableWeight())
	modHashSelector.Refresh(sortedEps)
	e.activeEp = sortedEps
	e.activeEpRoundRobin = roundRobinSelector
	e.activeEpConHash = conHashSelector
	e.activeEpModHash = modHashSelector
}

func (e *tarsEndpointManager) enableWeight() bool {
	return e.weightType == endpoint.EStaticWeight
}
