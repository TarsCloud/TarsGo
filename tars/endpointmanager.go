package tars

import (
	"encoding/json"
	"fmt"
	"hash/crc32"
	"io/ioutil"
	"math/rand"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/TarsCloud/TarsGo/tars/protocol/res/endpointf"
	"github.com/TarsCloud/TarsGo/tars/protocol/res/queryf"
	"github.com/TarsCloud/TarsGo/tars/util/consistenthash"
	"github.com/TarsCloud/TarsGo/tars/util/endpoint"
	"github.com/TarsCloud/TarsGo/tars/util/gtime"
)

//EndpointManager interface of naming system
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
func GetManager(comm *Communicator, objName string) EndpointManager {
	//taf
	initOnceGManager(comm.Client.RefreshEndpointInterval, comm.Client.CheckStatusInterval)
	g := gManager
	g.mlock.Lock()
	key := objName + comm.hashKey()
	if v, ok := g.eps[key]; ok {
		g.mlock.Unlock()
		return v
	}
	g.mlock.Unlock()

	TLOG.Debug("Create endpoint manager for ", objName)
	em := newTarsEndpointManager(objName, comm) // avoid dead lock
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
				em.activeEp = newEps[:] // replace ep list
				chmap := consistenthash.NewChMap(32)
				for _, e := range em.activeEp {
					chmap.Add(e)
				}
				em.activeEpHashMap = chmap
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
				TLOG.Errorf("update endoint error, %s.", e.objName)
			}

		}

		//cache to file
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
	directproxy bool
	comm        *Communicator
	locator     *queryf.QueryF

	epList      *sync.Map
	epLock      *sync.Mutex
	activeEp    []endpoint.Endpoint
	pos         int32
	activeEpf   []endpointf.EndpointF
	inactiveEpf []endpointf.EndpointF
	aliveCheck  chan endpointf.EndpointF

	checkAdapterList *sync.Map
	checkAdapter     chan *AdapterProxy

	activeEpHashMap *consistenthash.ChMap
	freshLock       *sync.Mutex
	lastInvoke      int64
	invokeNum       int32
}

func newTarsEndpointManager(objName string, comm *Communicator) *tarsEndpointManager {
	if objName == "" {
		return nil
	}
	e := &tarsEndpointManager{}
	e.comm = comm
	e.freshLock = &sync.Mutex{}
	e.epList = &sync.Map{}
	e.epLock = &sync.Mutex{}
	e.checkAdapterList = &sync.Map{}
	pos := strings.Index(objName, "@")
	if pos > 0 {
		//[direct]
		e.objName = objName[0:pos]
		endpoints := objName[pos+1:]
		e.directproxy = true
		ends := strings.Split(endpoints, ":")
		eps := make([]endpoint.Endpoint, len(ends))
		for i, end := range ends {
			eps[i] = endpoint.Parse(end)
		}
		e.activeEp = eps

		chmap := consistenthash.NewChMap(32)
		for _, e := range e.activeEp {
			chmap.Add(e)
		}
		e.activeEpHashMap = chmap
	} else {
		//[proxy] TODO singleton
		TLOG.Debug("proxy mode:", objName)
		e.objName = objName
		e.directproxy = false
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
	//only in active epf need to check.
	for _, ef := range e.activeEpf {
		ep := endpoint.Tars2endpoint(ef)
		if v, ok := e.epList.Load(ep.Key); ok {
			firstTime, needCheck := v.(*AdapterProxy).checkActive()
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

				e.activeEpHashMap.Remove(ep.Key)
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
		return crc32.ChecksumIEEE([]byte(sortedEps[i].Key)) < crc32.ChecksumIEEE([]byte(sortedEps[i].Key))
	})
	e.activeEp = sortedEps
	e.activeEpHashMap.Add(ep)
	e.epLock.Unlock()
}

// SelectAdapterProxy returns the selected adapter.
func (e *tarsEndpointManager) SelectAdapterProxy(msg *Message) (*AdapterProxy, bool) {
	e.epLock.Lock()
	eps := e.activeEp[:]
	e.epLock.Unlock()

	if e.directproxy && len(eps) == 0 {
		return nil, false
	}
	if !e.directproxy && len(e.activeEpf) == 0 {
		return nil, false
	}
	select {
	case adp := <-e.checkAdapter:
		TLOG.Errorf("SelectAdapterProxy|check adapter, ep: %+v", adp.GetPoint())
		e.checkAdapterList.Delete(endpoint.Tars2endpoint(*adp.GetPoint()).Key)
		return adp, true
	default:
	}
	var adp *AdapterProxy
	if msg.isHash && msg.hashType == ConsistentHash {
		if epi, ok := e.activeEpHashMap.FindUint32(uint32(msg.hashCode)); ok {
			ep := epi.(endpoint.Endpoint)
			if v, ok := e.epList.Load(ep.Key); ok {
				adp = v.(*AdapterProxy)
			} else {
				epf := endpoint.Endpoint2tars(ep)
				adp = NewAdapterProxy(&epf, e.comm)
				e.epList.Store(ep.Key, adp)
			}
		}
	} else {
		if len(eps) != 0 {
			var index int
			if msg.isHash && msg.hashType == ModHash {
				index = int(msg.hashCode) % len(eps)
			} else {
				e.pos = (e.pos + 1) % int32(len(eps))
				index = int(e.pos)
			}
			ep := eps[index]
			if v, ok := e.epList.Load(ep.Key); ok {
				adp = v.(*AdapterProxy)
			} else {
				epf := endpoint.Endpoint2tars(ep)
				adp = NewAdapterProxy(&epf, e.comm)
				e.epList.Store(ep.Key, adp)
			}
		}
	}
	if adp == nil && !e.directproxy {
		//No any node is alive ,just select a random one.
		randomIndex := rand.Intn(len(e.activeEpf))
		randomEpf := e.activeEpf[randomIndex]
		randomEp := endpoint.Tars2endpoint(randomEpf)
		if v, ok := e.epList.Load(randomEp.Key); ok {
			adp = v.(*AdapterProxy)
		} else {
			adp = NewAdapterProxy(&randomEpf, e.comm)
			e.epList.Store(randomEp.Key, adp)
		}
	}
	return adp, false
}

func (e *tarsEndpointManager) doFresh() error {
	if e.directproxy {
		return nil
	}
	e.freshLock.Lock()
	defer e.freshLock.Unlock()
	err := e.findAndSetObj(e.locator)
	return err
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
	var setable, ok bool
	var setID string
	var ret int32
	var err error
	if setable, ok = e.comm.GetPropertyBool("enableset"); ok {
		setID, _ = e.comm.GetProperty("setdivision")
	}
	if setable {
		ret, err = q.FindObjectByIdInSameSet(e.objName, setID, &activeEp, &inactiveEp)
	} else {
		ret, err = q.FindObjectByIdInSameGroup(e.objName, &activeEp, &inactiveEp)
	}
	if err != nil {
		TLOG.Errorf("findAndSetObj %s fail, error: %v", e.objName, err)
		return err
	}
	if ret != 0 {
		e := fmt.Errorf("findAndSetObj %s fail, ret: %d", e.objName, ret)
		TLOG.Error(e.Error())
		return e
	}

	// compare, assert in same order
	/*
		if endpoint.IsEqaul(activeEp, &e.activeEpf) {
			TLOG.Debug("endpoint not change: ", e.objName)
			return nil
		}
	*/
	if len(activeEp) == 0 {
		TLOG.Errorf("findAndSetObj %s, empty of active endpoint", e.objName)
		return nil
	}
	TLOG.Debugf("findAndSetObj|call FindObjectById ok, obj: %s, ret: %d, active: %v, inative: %v", e.objName, ret, activeEp, inactiveEp)

	newEps := make([]endpoint.Endpoint, len(activeEp))
	for i, ep := range activeEp {
		newEps[i] = endpoint.Tars2endpoint(ep)
	}

	//delete useless cache
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

	//delete active endpoint which status is false
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
	}

	//make endponit slice sorted
	sort.Slice(sortedEps, func(i int, j int) bool {
		return crc32.ChecksumIEEE([]byte(sortedEps[i].Key)) < crc32.ChecksumIEEE([]byte(sortedEps[i].Key))
	})

	chmap := consistenthash.NewChMap(32)
	for _, e := range sortedEps {
		chmap.Add(e)
	}

	e.epLock.Lock()
	e.activeEpf = activeEp
	e.inactiveEpf = inactiveEp
	e.activeEp = sortedEps
	e.activeEpHashMap = chmap
	e.epLock.Unlock()

	TLOG.Debugf("findAndSetObj|activeEp: %+v", sortedEps)
	return nil
}
