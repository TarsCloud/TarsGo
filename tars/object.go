package tars

import (
	"errors"
	"fmt"
	"github.com/TarsCloud/TarsGo/tars/protocol/res/requestf"
	"github.com/TarsCloud/TarsGo/tars/util/rtimer"
	"sync"
	"sync/atomic"
	"time"
)

type ObjectProxy struct {
	manager  *EndpointManager
	comm     *Communicator
	queueLen int32
}

func (obj *ObjectProxy) Init(comm *Communicator, objName string) {
	obj.comm = comm
	obj.manager = new(EndpointManager)
	obj.manager.Init(objName, obj.comm)
}

func (obj *ObjectProxy) Invoke(msg *Message, timeout time.Duration) error {
	adp := obj.manager.SelectAdapterProxy(msg)
	if adp == nil {
		TLOG.Error("no adapter Proxy selected:" + msg.Req.SServantName)
		return errors.New("no adapter Proxy selected:" + msg.Req.SServantName)
	}
	if obj.queueLen > ObjQueueMax {
		return errors.New("invoke queue is full:" + msg.Req.SServantName)
	}
	msg.Adp = adp
	atomic.AddInt32(&obj.queueLen, 1)
	readCh := make(chan *requestf.ResponsePacket, 1)
	adp.resp.Store(msg.Req.IRequestId, readCh)
	defer func() {
		checkPanic()
		atomic.AddInt32(&obj.queueLen, -1)
		adp.resp.Delete(msg.Req.IRequestId)
		close(readCh)
	}()
	if err := adp.Send(msg.Req); err != nil {
		return err
	}
	select {
	case <-rtimer.After(timeout):
		TLOG.Error("req timeout:", msg.Req.IRequestId)
		//TODO set resp ret from base.tars
		//msg.Resp.IRet = -1
		adp.failAdd()
		return errors.New(fmt.Sprintf("%s|%s|%d", "request timeout", msg.Req.SServantName, msg.Req.IRequestId))
	case msg.Resp = <-readCh:
		TLOG.Debug("recv msg succ ", msg.Req.IRequestId)
	}
	return nil
}

type ObjectProxyFactory struct {
	objs map[string]*ObjectProxy
	comm *Communicator
	om   *sync.Mutex
}

func (o *ObjectProxyFactory) Init(comm *Communicator) {
	o.om = new(sync.Mutex)
	o.comm = comm
	o.objs = make(map[string]*ObjectProxy)
}

func (o *ObjectProxyFactory) GetObjectProxy(objName string) *ObjectProxy {
	o.om.Lock()
	defer o.om.Unlock()
	if obj, ok := o.objs[objName]; ok {
		return obj
	}
	obj := new(ObjectProxy)
	obj.Init(o.comm, objName)
	o.objs[objName] = obj
	return obj
}
