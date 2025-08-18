package tars

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/TarsCloud/TarsGo/tars/model"
	"github.com/TarsCloud/TarsGo/tars/protocol/res/basef"
	"github.com/TarsCloud/TarsGo/tars/protocol/res/endpointf"
	"github.com/TarsCloud/TarsGo/tars/protocol/res/requestf"
	"github.com/TarsCloud/TarsGo/tars/transport"
	"github.com/TarsCloud/TarsGo/tars/util/endpoint"
	"github.com/TarsCloud/TarsGo/tars/util/rtimer"
	"github.com/TarsCloud/TarsGo/tars/util/tools"
)

// AdapterProxy : Adapter proxy
type AdapterProxy struct {
	resp              sync.Map
	point             *endpointf.EndpointF
	tarsClient        *transport.TarsClient
	conf              *transport.TarsClientConf
	comm              *Communicator
	servantProxy      *ServantProxy
	failCount         int32
	lastFailCount     int32
	sendCount         int32
	successCount      int32
	status            bool // true for good
	lastSuccessTime   int64
	lastBlockTime     int64
	lastCheckTime     int64
	lastKeepAliveTime int64
	pushCallback      func([]byte)
	onceKeepAlive     sync.Once
	onDispatch        model.Ondispatch
	asyncPushCh       chan *requestf.ResponsePacket //
	closed            bool
}

// NewAdapterProxy create an adapter proxy
func NewAdapterProxy(objName string, point *endpointf.EndpointF, comm *Communicator) *AdapterProxy {
	c := &AdapterProxy{}
	if c.asyncPushCh == nil {
		c.asyncPushCh = make(chan *requestf.ResponsePacket)
	}
	c.comm = comm
	c.point = point
	proto := "tcp"
	if point.Istcp == endpoint.UDP {
		proto = "udp"
	} else if point.Istcp == endpoint.SSL {
		proto = "ssl"
	}
	conf := &transport.TarsClientConf{
		Proto:        proto,
		QueueLen:     comm.Client.ClientQueueLen,
		IdleTimeout:  comm.Client.ClientIdleTimeout,
		ReadTimeout:  comm.Client.ClientReadTimeout,
		WriteTimeout: comm.Client.ClientWriteTimeout,
		DialTimeout:  comm.Client.ClientDialTimeout,
	}
	if point.Istcp == endpoint.SSL {
		if tlsConfig, ok := comm.app.clientObjTlsConfig[objName]; ok {
			conf.TlsConfig = tlsConfig
		} else {
			conf.TlsConfig = comm.app.clientTlsConfig
		}
	}
	c.conf = conf
	c.tarsClient = transport.NewTarsClient(fmt.Sprintf("%s:%d", point.Host, point.Port), c, conf)
	c.status = true

	// push queue listenning
	go func() {
		for {
			select {
			case resp := <-c.asyncPushCh:
				c.onPush(resp)
			}
		}
	}()
	return c
}

// ParsePackage : Parse packet from bytes
func (c *AdapterProxy) ParsePackage(buff []byte) (int, int) {
	return c.servantProxy.proto.ParsePackage(buff)
}

// Recv : Recover read channel when closed for timeout
func (c *AdapterProxy) Recv(pkg []byte) {
	defer func() {
		// TODO readCh has a certain probability to be closed after the load, and we need to recover
		// Maybe there is a better way
		if err := recover(); err != nil {
			TLOG.Error("recv pkg panic:", err)
		}
	}()
	packet, err := c.servantProxy.proto.ResponseUnpack(pkg)
	if err != nil {
		TLOG.Errorf("decode packet error: %v", err)
		return
	}
	if packet.IRequestId == 0 {
		c.asyncPushCh <- packet
		return
	}
	if packet.CPacketType == basef.TARSONEWAY {
		return
	}
	chIF, ok := c.resp.Load(packet.IRequestId)
	if ok {
		ch := chIF.(chan *requestf.ResponsePacket)
		select {
		case ch <- packet:
		// after conf.ReadTimeout, release this goroutine to make sure response package is received by Tars_Invoke().
		case <-rtimer.After(c.conf.ReadTimeout):
			TLOG.Errorf("response timeout, write channel error, now time :%v, RequestId:%v",
				time.Now().UnixNano()/1e6, packet.IRequestId)
		}
	} else {
		TLOG.Errorf("response timeout, req has been drop, now time :%v, RequestId:%v",
			time.Now().UnixNano()/1e6, packet.IRequestId)
	}
}

// Send : Send packet
func (c *AdapterProxy) Send(req *requestf.RequestPacket) error {
	TLOG.Debug("send req:", req.IRequestId)
	c.sendAdd()
	sbuf, err := c.servantProxy.proto.RequestPack(req)
	if err != nil {
		TLOG.Debug("protocol wrong:", req.IRequestId)
		return err
	}
	return c.tarsClient.Send(sbuf)
}

// GetPoint get an endpoint
func (c *AdapterProxy) GetPoint() *endpointf.EndpointF {
	return c.point
}

// Close the client
func (c *AdapterProxy) Close() {
	c.tarsClient.Close()
	c.closed = true
}

func (c *AdapterProxy) sendAdd() {
	atomic.AddInt32(&c.sendCount, 1)
}

func (c *AdapterProxy) successAdd() {
	now := time.Now().Unix()
	atomic.SwapInt64(&c.lastSuccessTime, now)
	atomic.AddInt32(&c.successCount, 1)
	atomic.SwapInt32(&c.lastFailCount, 0)
}

func (c *AdapterProxy) failAdd() {
	atomic.AddInt32(&c.lastFailCount, 1)
	atomic.AddInt32(&c.failCount, 1)
}

func (c *AdapterProxy) reset() {
	now := time.Now().Unix()
	atomic.SwapInt32(&c.sendCount, 0)
	atomic.SwapInt32(&c.successCount, 0)
	atomic.SwapInt32(&c.failCount, 0)
	atomic.SwapInt32(&c.lastFailCount, 0)
	atomic.SwapInt64(&c.lastBlockTime, now)
	atomic.SwapInt64(&c.lastCheckTime, now)
	atomic.SwapInt64(&c.lastKeepAliveTime, now)
	c.status = true
}

func (c *AdapterProxy) checkActive() (firstTime bool, needCheck bool) {
	if c.closed {
		return false, false
	}

	now := time.Now().Unix()
	if c.status {
		//check if healthy，fail 5 times in a row within 5s
		if (now-c.lastSuccessTime) >= failInterval && c.lastFailCount >= fainN {
			c.status = false
			c.lastBlockTime = now
			return true, false
		}
		// lgt 60s
		if (now - c.lastCheckTime) >= checkTime {
			c.lastBlockTime = now
			// The number of failures is more than 2 and the failure rate is more than 50%
			if c.failCount >= overN && (float32(c.failCount)/float32(c.sendCount)) >= failRatio {
				c.status = false
				return true, false
			}
			return false, false
		}
		return false, false
	}

	// lgt 30s
	if (now - c.lastBlockTime) >= tryTimeInterval {
		c.lastBlockTime = now
		if err := c.tarsClient.ReConnect(); err != nil {
			return false, false
		}

		return false, true
	}

	return false, false
}

func (c *AdapterProxy) onPush(pkg *requestf.ResponsePacket) {
	if pkg.SResultDesc == reconnectMsg {
		TLOG.Infof("reconnect %s:%d", c.point.Host, c.point.Port)
		oldClient := c.tarsClient
		c.tarsClient = transport.NewTarsClient(fmt.Sprintf("%s:%d", c.point.Host, c.point.Port), c, c.conf)

		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*ClientIdleTimeout)
		defer cancel()
		oldClient.GraceClose(ctx) // grace shutdown
		return
	}

	switch pkg.IVersion {
	case 0:
		// raw socket push
		if c.pushCallback == nil {
			return
		}
		data := tools.Int8ToByte(pkg.SBuffer)
		c.pushCallback(data)

	case basef.TARSVERSION:
		if c.onDispatch == nil {
			return
		}
		// tars proto push
		c.onDispatch.Ondispatch(pkg)
	}
}

func (c *AdapterProxy) autoKeepAlive() {
	interval := c.comm.Client.ClientIdleTimeout / 2
	if interval == 0 {
		interval = time.Minute
	}
	for range time.NewTicker(interval).C {
		if c.closed {
			return
		}
		c.doKeepAlive()
	}
}

func (c *AdapterProxy) OnConnect(address string) {
	if c.servantProxy.onConnectCallback != nil {
		c.servantProxy.onConnectCallback(address)
	}
}

func (c *AdapterProxy) OnClose(address string) {
	if c.servantProxy.onCloseCallback != nil {
		c.servantProxy.onCloseCallback(address)
	}
}

func (c *AdapterProxy) doKeepAlive() {
	if c.closed {
		return
	}

	if c.servantProxy.queueLen > c.comm.Client.ObjQueueMax {
		return
	}

	now := time.Now().Unix()
	if now-c.lastKeepAliveTime < int64(c.comm.Client.KeepAliveInterval/1000) {
		return
	}
	c.lastKeepAliveTime = now

	req := requestf.RequestPacket{
		IVersion:     c.servantProxy.version,
		CPacketType:  basef.TARSONEWAY,
		IRequestId:   c.servantProxy.genRequestID(),
		SServantName: c.servantProxy.name,
		SFuncName:    "tars_ping",
		ITimeout:     int32(c.servantProxy.timeout),
	}
	msg := &Message{Req: &req, Ser: c.servantProxy}
	msg.Init()

	msg.Adp = c
	atomic.AddInt32(&c.servantProxy.queueLen, 1)
	defer func() {
		CheckPanic()
		atomic.AddInt32(&c.servantProxy.queueLen, -1)
	}()
	if err := c.Send(msg.Req); err != nil {
		c.failAdd()
		return
	}
	c.successAdd()
}
