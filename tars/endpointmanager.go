package tars

import (
	"github.com/TarsCloud/TarsGo/tars/protocol/res/endpointf"
	"github.com/TarsCloud/TarsGo/tars/protocol/res/queryf"
	"github.com/TarsCloud/TarsGo/tars/util/endpoint"
	"github.com/TarsCloud/TarsGo/tars/util/set"
	"strings"
	"sync"
	"time"
)

type EndpointManager struct {
	objName         string
	directproxy     bool
	adapters        map[endpoint.Endpoint]*AdapterProxy
	index           []interface{} //cache the set
	pointsSet       *set.Set
	comm            *Communicator
	mlock           *sync.Mutex
	refreshInterval int
	pos             int32
	depth           int32
}

func (e *EndpointManager) setObjName(objName string) {
	if objName == "" {
		return
	}
	pos := strings.Index(objName, "@")
	if pos > 0 {
		//[direct]
		e.objName = objName[0:pos]
		endpoints := objName[pos+1:]
		e.directproxy = true
		for _, end := range strings.Split(endpoints, ":") {
			e.pointsSet.Add(endpoint.Parse(end))
		}
		e.index = e.pointsSet.Slice()

	} else {
		//[proxy] TODO singleton
		TLOG.Debug("proxy mode:", objName)
		e.objName = objName
		//comm := NewCommunicator()
		//comm.SetProperty("netthread", 1)
		obj, _ := e.comm.GetProperty("locator")
		q := new(queryf.QueryF)
		e.comm.StringToProxy(obj, q)
		e.findAndSetObj(q)
		go func() {
			loop := time.NewTicker(time.Duration(e.refreshInterval) * time.Millisecond)
			for range loop.C {
				//TODO exit
				e.findAndSetObj(q)
			}
		}()
	}
}

func (e *EndpointManager) Init(objName string, comm *Communicator) error {
	e.comm = comm
	e.mlock = new(sync.Mutex)
	e.adapters = make(map[endpoint.Endpoint]*AdapterProxy)
	e.pointsSet = set.NewSet()
	e.directproxy = false
	e.refreshInterval = comm.Client.refreshEndpointInterval
	e.pos = 0
	e.depth = 0
	//ObjName要放到最后初始化
	e.setObjName(objName)
	return nil
}

func (e *EndpointManager) GetNextValidProxy() *AdapterProxy {
	e.mlock.Lock()
	defer e.mlock.Unlock()
	ep := e.GetNextEndpoint()
	if ep == nil {
		return nil
	}
	if adp, ok := e.adapters[*ep]; ok {
		//如果递归了所有节点还没有找到可用节点就返回nil
		if adp.status {
			return adp
		} else if e.depth > e.pointsSet.Len() {
			return nil
		} else {
			e.depth++
			return e.GetNextValidProxy()
		}
	}
	err := e.createProxy(*ep)
	if err != nil {
		TLOG.Error("create adapter fail:", *ep, err)
		return nil
	}
	return e.adapters[*ep]
}

func (e *EndpointManager) GetNextEndpoint() *endpoint.Endpoint {
	length := len(e.index)
	if length <= 0 {
		return nil
	}
	var ep endpoint.Endpoint
	e.pos = (e.pos + 1) % int32(length)
	ep = e.index[e.pos].(endpoint.Endpoint)
	return &ep
}

func (e *EndpointManager) createProxy(ep endpoint.Endpoint) error {
	TLOG.Debug("create adapter:", ep)
	adp := new(AdapterProxy)
	//TODO
	end := endpoint.Endpoint2tars(ep)
	err := adp.New(&end, e.comm)
	if err != nil {
		return err
	}
	e.adapters[ep] = adp
	return nil
}

func (e *EndpointManager) GetHashProxy(hashcode int64) *AdapterProxy {
	//非常不安全的hash
	ep := e.GetHashEndpoint(hashcode)
	if ep == nil {
		return nil
	}
	if adp, ok := e.adapters[*ep]; ok {
		return adp
	}
	err := e.createProxy(*ep)
	if err != nil {
		TLOG.Error("create adapter fail:", ep, err)
		return nil
	}
	return e.adapters[*ep]
}

func (e *EndpointManager) GetHashEndpoint(hashcode int64) *endpoint.Endpoint {
	length := len(e.index)
	if length <= 0 {
		return nil
	}
	pos := hashcode % int64(length)
	ep := e.index[pos].(endpoint.Endpoint)
	return &ep
}

func (e *EndpointManager) SelectAdapterProxy(msg *Message) *AdapterProxy {
	if msg.isHash {
		return e.GetHashProxy(msg.hashCode)
	}
	return e.GetNextValidProxy()
}

func (e *EndpointManager) findAndSetObj(q *queryf.QueryF) {
	activeEp := new([]endpointf.EndpointF)
	inactiveEp := new([]endpointf.EndpointF)
	var setable, ok bool
	var setId string
	var ret int32
	var err error
	if setable, ok = e.comm.GetPropertyBool("enableset"); ok {
		setId, _ = e.comm.GetProperty("setdivision")
	}
	if setable {
		ret, err = q.FindObjectByIdInSameSet(e.objName, setId, activeEp, inactiveEp)
	} else {
		ret, err = q.FindObjectByIdInSameGroup(e.objName, activeEp, inactiveEp)
	}
	if err != nil {
		TLOG.Error("find obj end fail:", err.Error())
		return
	}
	TLOG.Debug("find obj endpoint:", e.objName, ret, *activeEp, *inactiveEp)

	e.mlock.Lock()
	if (len(*inactiveEp)) > 0 {
		for _, ep := range *inactiveEp {
			end := endpoint.Tars2endpoint(ep)
			e.pointsSet.Remove(end)
			if a, ok := e.adapters[end]; ok {
				delete(e.adapters, end)
				a.Close()
			}
		}
	}
	if (len(*activeEp)) > 0 {
		e.pointsSet.Clear() //先清空，再加回去，这里导致必须加锁，不清又可能导致泄漏，以后改成remove元素会好
		for _, ep := range *activeEp {
			end := endpoint.Tars2endpoint(ep)
			e.pointsSet.Add(end)
		}
		e.index = e.pointsSet.Slice()
	}
	for end, _ := range e.adapters {
		//清理掉脏数据
		if !e.pointsSet.Has(end) {
			if a, ok := e.adapters[end]; ok {
				delete(e.adapters, end)
				a.Close()
			}
		}
	}
	e.mlock.Unlock()
}
