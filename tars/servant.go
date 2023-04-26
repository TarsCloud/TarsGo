package tars

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/TarsCloud/TarsGo/tars/model"
	"github.com/TarsCloud/TarsGo/tars/protocol"
	"github.com/TarsCloud/TarsGo/tars/protocol/res/basef"
	"github.com/TarsCloud/TarsGo/tars/protocol/res/requestf"
	"github.com/TarsCloud/TarsGo/tars/util/current"
	"github.com/TarsCloud/TarsGo/tars/util/endpoint"
	"github.com/TarsCloud/TarsGo/tars/util/rtimer"
)

var (
	maxInt32 int32 = 1<<31 - 1
	msgID    int32
	_        model.Servant = (*ServantProxy)(nil)
)

const (
	StatSuccess = iota
	StatFailed
)

// ServantProxy tars servant proxy instance
type ServantProxy struct {
	name         string
	comm         *Communicator
	manager      EndpointManager
	syncTimeout  int
	asyncTimeout int
	version      int16
	proto        model.Protocol
	queueLen     int32

	pushCallback func([]byte)
}

// NewServantProxy creates and initializes a servant proxy
func NewServantProxy(comm *Communicator, objName string, opts ...EndpointManagerOption) *ServantProxy {
	return newServantProxy(comm, objName, opts...)
}

func newServantProxy(comm *Communicator, objName string, opts ...EndpointManagerOption) *ServantProxy {
	s := &ServantProxy{
		comm:    comm,
		proto:   &protocol.TarsProtocol{},
		version: basef.TARSVERSION,
	}
	pos := strings.Index(objName, "@")
	if pos > 0 {
		s.name = objName[0:pos]
	} else {
		s.name = objName
	}
	pos = strings.Index(s.name, "://")
	if pos > 0 {
		s.name = s.name[pos+3:]
	}

	// init manager
	s.manager = GetManager(comm, objName, opts...)

	s.comm = comm
	s.proto = &protocol.TarsProtocol{}
	s.syncTimeout = s.comm.Client.SyncInvokeTimeout
	s.asyncTimeout = s.comm.Client.AsyncInvokeTimeout
	s.version = basef.TARSVERSION
	return s
}

// Name is obj name
func (s *ServantProxy) Name() string {
	return s.name
}

// TarsSetTimeout sets the timeout for client calling the server , which is in ms.
func (s *ServantProxy) TarsSetTimeout(t int) {
	s.syncTimeout = t
}

// TarsSetVersion set tars version
func (s *ServantProxy) TarsSetVersion(iVersion int16) {
	s.version = iVersion
}

// TarsSetProtocol tars set model protocol
func (s *ServantProxy) TarsSetProtocol(proto model.Protocol) {
	s.proto = proto
}

// Endpoints returns all active endpoint.Endpoint
func (s *ServantProxy) Endpoints() []*endpoint.Endpoint {
	return s.manager.GetAllEndpoint()
}

// 生成请求 ID
func (s *ServantProxy) genRequestID() int32 {
	// 尽力防止溢出
	atomic.CompareAndSwapInt32(&msgID, maxInt32, 1)
	for {
		// 0比较特殊,用于表示 server 端推送消息给 client 端进行主动 close()
		// 溢出后回转成负数
		if v := atomic.AddInt32(&msgID, 1); v != 0 {
			return v
		}
	}
}

// SetPushCallback set callback function for pushing
func (s *ServantProxy) SetPushCallback(callback func([]byte)) {
	s.pushCallback = callback
}

// TarsInvoke is used for client invoking server.
func (s *ServantProxy) TarsInvoke(ctx context.Context, cType byte,
	sFuncName string,
	buf []byte,
	status map[string]string,
	reqContext map[string]string,
	resp *requestf.ResponsePacket) error {
	defer CheckPanic()

	msg := buildMessage(ctx, cType, sFuncName, buf, status, reqContext, resp, s)
	timeout := time.Duration(s.syncTimeout) * time.Millisecond
	err := s.invokeFilters(ctx, msg, timeout)

	if err != nil {
		return err
	}
	*resp = *msg.Resp
	return nil
}

// TarsInvokeAsync is used for client invoking server.
func (s *ServantProxy) TarsInvokeAsync(ctx context.Context, cType byte,
	sFuncName string,
	buf []byte,
	status map[string]string,
	reqContext map[string]string,
	resp *requestf.ResponsePacket,
	callback model.Callback) error {
	defer CheckPanic()

	msg := buildMessage(ctx, cType, sFuncName, buf, status, reqContext, resp, s)
	msg.Req.ITimeout = int32(s.asyncTimeout)
	if callback == nil {
		msg.Req.CPacketType = basef.TARSONEWAY
	} else {
		msg.Async = true
		msg.Callback = callback
	}

	timeout := time.Duration(s.asyncTimeout) * time.Millisecond
	return s.invokeFilters(ctx, msg, timeout)
}

func (s *ServantProxy) invokeFilters(ctx context.Context, msg *Message, timeout time.Duration) error {
	if ok, to, isTimeout := current.GetClientTimeout(ctx); ok && isTimeout {
		timeout = time.Duration(to) * time.Millisecond
		msg.Req.ITimeout = int32(to)
	}

	var err error
	s.manager.preInvoke()
	app := s.comm.app
	if app.allFilters.cf != nil {
		err = app.allFilters.cf(ctx, msg, s.doInvoke, timeout)
	} else if cf := app.getMiddlewareClientFilter(); cf != nil {
		err = cf(ctx, msg, s.doInvoke, timeout)
	} else {
		// execute pre client filters
		for i, v := range app.allFilters.preCfs {
			err = v(ctx, msg, s.doInvoke, timeout)
			if err != nil {
				TLOG.Errorf("Pre filter error, no: %v, err: %v", i, err.Error())
			}
		}
		// execute rpc
		err = s.doInvoke(ctx, msg, timeout)
		// execute post client filters
		for i, v := range app.allFilters.postCfs {
			filterErr := v(ctx, msg, s.doInvoke, timeout)
			if filterErr != nil {
				TLOG.Errorf("Post filter error, no: %v, err: %v", i, filterErr.Error())
			}
		}
	}
	// no async rpc call
	if !msg.Async {
		s.manager.postInvoke()
		msg.End()
		s.reportStat(msg, err)
	}

	return err
}

func (s *ServantProxy) reportStat(msg *Message, err error) {
	if err != nil {
		TLOG.Errorf("Invoke error: %s, %s, %v, cost:%d", s.name, msg.Req.SFuncName, err.Error(), msg.Cost())
		if msg.Resp == nil {
			ReportStat(msg, StatSuccess, StatSuccess, StatFailed)
		} else if msg.Status == basef.TARSINVOKETIMEOUT {
			ReportStat(msg, StatSuccess, StatFailed, StatSuccess)
		} else {
			ReportStat(msg, StatSuccess, StatSuccess, StatFailed)
		}
		return
	}
	ReportStat(msg, StatFailed, StatSuccess, StatSuccess)
}

func (s *ServantProxy) doInvoke(ctx context.Context, msg *Message, timeout time.Duration) (err error) {
	adp, needCheck := s.manager.SelectAdapterProxy(msg)
	if adp == nil {
		return errors.New("no adapter Proxy selected:" + msg.Req.SServantName)
	}
	if s.queueLen > adp.comm.Client.ObjQueueMax {
		return errors.New("invoke queue is full:" + msg.Req.SServantName)
	}
	ep := adp.GetPoint()
	current.SetServerIPWithContext(ctx, ep.Host)
	current.SetServerPortWithContext(ctx, fmt.Sprintf("%v", ep.Port))
	msg.Adp = adp
	adp.servantProxy = s

	if s.pushCallback != nil {
		// auto keep alive for push client
		go adp.onceKeepAlive.Do(adp.autoKeepAlive)
		adp.pushCallback = s.pushCallback
	}

	atomic.AddInt32(&s.queueLen, 1)
	readCh := make(chan *requestf.ResponsePacket)
	adp.resp.Store(msg.Req.IRequestId, readCh)
	var releaseFunc = func() {
		CheckPanic()
		atomic.AddInt32(&s.queueLen, -1)
		adp.resp.Delete(msg.Req.IRequestId)
	}
	defer func() {
		if !msg.Async || err != nil {
			releaseFunc()
		}
	}()

	if err = adp.Send(msg.Req); err != nil {
		adp.failAdd()
		return err
	}

	if msg.Req.CPacketType == basef.TARSONEWAY {
		adp.successAdd()
		return nil
	}

	// async call rpc
	if msg.Async {
		go func() {
			defer releaseFunc()
			err := s.waitInvoke(msg, adp, timeout, needCheck)
			s.manager.postInvoke()
			msg.End()
			s.reportStat(msg, err)
			if msg.Status != basef.TARSINVOKETIMEOUT {
				current.SetResponseContext(ctx, msg.Resp.Context)
				current.SetResponseStatus(ctx, msg.Resp.Status)
			}
			if _, err := msg.Callback.Dispatch(ctx, msg.Req, msg.Resp, err); err != nil {
				TLOG.Errorf("Callback error: %s, %s, %+v", s.name, msg.Req.SFuncName, err)
			}
		}()
		return nil
	}

	return s.waitInvoke(msg, adp, timeout, needCheck)
}

func (s *ServantProxy) waitInvoke(msg *Message, adp *AdapterProxy, timeout time.Duration, needCheck bool) error {
	ch, _ := adp.resp.Load(msg.Req.IRequestId)
	readCh := ch.(chan *requestf.ResponsePacket)

	select {
	case <-rtimer.After(timeout):
		msg.Status = basef.TARSINVOKETIMEOUT
		adp.failAdd()
		msg.End()
		return fmt.Errorf("request timeout, begin time:%d, cost:%d, obj:%s, func:%s, addr:(%s:%d), reqid:%d",
			msg.BeginTime, msg.Cost(), msg.Req.SServantName, msg.Req.SFuncName, adp.point.Host, adp.point.Port, msg.Req.IRequestId)
	case msg.Resp = <-readCh:
		if needCheck {
			go func() {
				adp.reset()
				ep := endpoint.Tars2endpoint(*msg.Adp.point)
				s.manager.addAliveEp(ep)
			}()
		}
		adp.successAdd()
		if msg.Resp != nil {
			if msg.Status != basef.TARSSERVERSUCCESS || msg.Resp.IRet != 0 {
				if msg.Resp.SResultDesc == "" {
					return fmt.Errorf("basef error code %d", msg.Resp.IRet)
				}
				if msg.Resp.IRet != 0 && msg.Resp.IRet != 1 {
					return &Error{Code: msg.Resp.IRet, Message: msg.Resp.SResultDesc}
				}
				return errors.New(msg.Resp.SResultDesc)
			}
		} else {
			TLOG.Debug("recv nil Resp, close of the readCh?")
		}
		TLOG.Debug("recv msg success ", msg.Req.IRequestId)
	}
	return nil
}
