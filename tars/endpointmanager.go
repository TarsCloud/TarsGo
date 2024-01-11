package tars

import (
	"context"
	"encoding/json"
	"hash/crc32"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/TarsCloud/TarsGo/tars/protocol/res/endpointf"
	"github.com/TarsCloud/TarsGo/tars/protocol/res/queryf"
	"github.com/TarsCloud/TarsGo/tars/registry"
	tarsregistry "github.com/TarsCloud/TarsGo/tars/registry/tars"
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
	eps                 map[string]*endpointManager
	mlock               *sync.Mutex
	app                 *application
	refreshInterval     int
	checkStatusInterval int
}

func initOnceGManager(app *application) {
	gManagerInitOnce.Do(func() {
		cltCfg := app.ClientConfig()
		gManager = &globalManager{app: app, refreshInterval: cltCfg.RefreshEndpointInterval, checkStatusInterval: cltCfg.CheckStatusInterval}
		gManager.eps = make(map[string]*endpointManager)
		gManager.mlock = &sync.Mutex{}
		go gManager.updateEndpoints()
		go gManager.checkEpStatus()
	})
}

// GetManager return a endpoint manager from global endpoint manager
func GetManager(comm *Communicator, objName string, opts ...EndpointManagerOption) EndpointManager {
	// tars
	initOnceGManager(comm.app)
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
	em := newEndpointManager(objName, comm, opts...) // avoid dead lock
	g.mlock.Lock()
	if v, ok := g.eps[key]; ok {
		g.mlock.Unlock()
		return v
	}
	g.eps[key] = em
	// if fresh is error,we should get it from cache
	if err := em.doFresh(); err != nil {
		for _, cache := range comm.app.appCache.ObjCaches {
			if em.objName == cache.Name && em.setDivision == cache.SetID && comm.GetLocator() == cache.Locator {
				em.activeEpf = cache.Endpoints
				newEps := make([]endpoint.Endpoint, len(em.activeEpf))
				for i, ep := range em.activeEpf {
					newEps[i] = endpoint.Tars2endpoint(ep)
				}
				em.updateActiveEp(newEps)
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
		eps := make([]*endpointManager, 0)
		for _, v := range g.eps {
			if v.registrar != nil {
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
		eps := make([]*endpointManager, 0)
		for _, v := range g.eps {
			if v.registrar != nil {
				eps = append(eps, v)
			}
		}
		g.mlock.Unlock()
		TLOG.Debugf("start refresh %d endpoints %d", len(eps), g.refreshInterval)
		for _, em := range eps {
			if err := em.doFresh(); err != nil {
				TLOG.Errorf("obj: %s update endpoint error: %v.", em.objName, err)
			}
		}

		// cache to file
		svrCfg := g.app.ServerConfig()
		if svrCfg != nil && svrCfg.DataPath != "" {
			cachePath := filepath.Join(svrCfg.DataPath, svrCfg.Server) + ".tarsdat"
			g.app.appCache.ModifyTime = gtime.CurrDateTime
			objCache := make([]ObjCache, len(eps))
			for i, e := range eps {
				objCache[i].Name = e.objName
				objCache[i].SetID = e.setDivision
				objCache[i].Locator = e.comm.GetLocator()
				objCache[i].Endpoints = e.activeEpf
				objCache[i].InactiveEndpoints = e.inactiveEpf
			}
			g.app.appCache.ObjCaches = objCache
			data, _ := json.MarshalIndent(&g.app.appCache, "", "    ")
			if err := os.WriteFile(cachePath, data, 0644); err != nil {
				TLOG.Errorf("update appCache error: %v", err)
			}
		}
	}
}

// endpointManager is a struct which contains endpoint information.
type endpointManager struct {
	objName     string // name only, no ip list
	enableSet   bool
	setDivision string
	directProxy bool
	comm        *Communicator
	registrar   registry.Registrar

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
	apply(e *endpointManager)
	key(k *string)
}

type OptionFunc struct {
	applyFunc func(*endpointManager)
	keyFunc   func(*string)
}

func (f OptionFunc) apply(e *endpointManager) {
	if f.applyFunc != nil {
		f.applyFunc(e)
	}
}

func (f OptionFunc) key(e *string) {
	if f.keyFunc != nil {
		f.keyFunc(e)
	}
}

func newOptionFunc(applyFunc func(*endpointManager), keyFunc func(*string)) OptionFunc {
	return OptionFunc{applyFunc: applyFunc, keyFunc: keyFunc}
}

func WithSet(setDivision string) OptionFunc {
	return newOptionFunc(func(e *endpointManager) {
		if setDivision != "" {
			e.enableSet = true
			e.setDivision = setDivision
		}
	}, func(s *string) {
		*s = *s + ":" + setDivision
	})
}

func newEndpointManager(objName string, comm *Communicator, opts ...EndpointManagerOption) *endpointManager {
	if objName == "" {
		return nil
	}
	e := &endpointManager{}
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
		e.updateActiveEp(eps)
	} else {
		// [proxy] TODO singleton
		TLOG.Debug("proxy mode:", objName)
		e.objName = objName
		e.directProxy = false
		if e.comm.opt.registrar == nil {
			obj, _ := e.comm.GetProperty("locator")
			query := new(queryf.QueryF)
			TLOG.Debug("string to proxy locator ", obj)
			e.comm.StringToProxy(obj, query)
			e.registrar = tarsregistry.New(query, e.comm.Client)
		} else {
			e.registrar = e.comm.opt.registrar
		}
		e.checkAdapter = make(chan *AdapterProxy, 1000)
	}
	return e
}

// GetAllEndpoint returns all endpoint information as a array(support not tars service).
func (e *endpointManager) GetAllEndpoint() []*endpoint.Endpoint {
	eps := e.activeEp[:]
	out := make([]*endpoint.Endpoint, len(eps))
	for i := 0; i < len(eps); i++ {
		out[i] = &eps[i]
	}
	return out
}

func (e *endpointManager) checkStatus() {
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
				if _, ok = e.checkAdapterList.Load(ep.Key); !ok {
					adp = v.(*AdapterProxy)
					e.checkAdapterList.Store(ep.Key, adp)
					e.checkAdapter <- adp
					TLOG.Errorf("checkStatus|insert check adapter, ep: %+v", ep.Key)
				}
			}
		}
	}
}

func (e *endpointManager) addAliveEp(ep endpoint.Endpoint) {
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
func (e *endpointManager) SelectAdapterProxy(msg *Message) (*AdapterProxy, bool) {
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
		TLOG.Errorf("SelectAdapterProxy|enableWeight: %v, isHash: %b, hashType: %s, hashCode: %d, err: %v", e.enableWeight(), msg.isHash, msg.hashType, msg.hashCode, err)
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

func (e *endpointManager) doFresh() error {
	if e.directProxy {
		return nil
	}
	e.freshLock.Lock()
	defer e.freshLock.Unlock()
	return e.refreshEndpoints()
}

func (e *endpointManager) preInvoke() {
	atomic.AddInt32(&e.invokeNum, 1)
	e.lastInvoke = gtime.CurrUnixTime
}

func (e *endpointManager) postInvoke() {
	atomic.AddInt32(&e.invokeNum, -1)
}

func (e *endpointManager) refreshEndpoints() error {
	var (
		activeEp, inactiveEp []endpointf.EndpointF
		enableSet, ok        bool
		setDivision          string
		err                  error
	)
	if e.enableSet && e.setDivision != "" {
		enableSet, setDivision = e.enableSet, e.setDivision
	} else if enableSet, ok = e.comm.GetPropertyBool("enableset"); ok {
		setDivision, _ = e.comm.GetProperty("setdivision")
	}

	if enableSet {
		activeEp, inactiveEp, err = e.registrar.QueryServantBySet(context.Background(), e.objName, setDivision)
	} else {
		activeEp, inactiveEp, err = e.registrar.QueryServant(context.Background(), e.objName)
	}
	if err != nil {
		return err
	}

	// sort activeEp slice
	sort.Slice(activeEp, func(i, j int) bool {
		return activeEp[i].Host < activeEp[j].Host
	})
	if reflect.DeepEqual(&activeEp, &e.activeEpf) {
		TLOG.Debugf("endpoint not change: %s, set: %s", e.objName, setDivision)
		return nil
	}

	if len(activeEp) == 0 {
		TLOG.Errorf("refreshEndpoints %s, empty of active endpoint", e.objName)
		return nil
	}
	e.epLock.Lock()
	e.activeEpf = activeEp
	e.inactiveEpf = inactiveEp
	e.epLock.Unlock()
	TLOG.Debugf("refreshEndpoints|call QueryServant or QueryServantBySet, obj: %s, set: %s, active: %v, inactive: %v", e.objName, setDivision, activeEp, inactiveEp)

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
			TLOG.Debugf("refreshEndpoints|delete useless endpoint from epList: %+v", key)
		}
		return true
	})

	e.updateActiveEp(newEps)
	return nil
}

func (e *endpointManager) updateActiveEp(newEps []endpoint.Endpoint) {
	if len(newEps) == 0 {
		return
	}
	sameType, lastType := true, newEps[0].WeightType
	// delete active endpoint which status is false
	sortedEps := make([]endpoint.Endpoint, 0, len(newEps))
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
			sameType = false
		}
	}

	e.weightType = endpoint.ELoop
	if sameType {
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

	e.epLock.Lock()
	e.activeEp = sortedEps
	e.activeEpRoundRobin = roundRobinSelector
	e.activeEpConHash = conHashSelector
	e.activeEpModHash = modHashSelector
	e.epLock.Unlock()

	TLOG.Debugf("updateActiveEp|activeEp: %+v", sortedEps)
}

func (e *endpointManager) enableWeight() bool {
	return e.weightType == endpoint.EStaticWeight
}
