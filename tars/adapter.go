package tars

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/TarsCloud/TarsGo/tars/protocol/res/basef"
	"github.com/TarsCloud/TarsGo/tars/protocol/res/endpointf"
	"github.com/TarsCloud/TarsGo/tars/protocol/res/requestf"
	"github.com/TarsCloud/TarsGo/tars/transport"
	"github.com/TarsCloud/TarsGo/tars/util/rtimer"
)

var reconnectMsg = "_reconnect_"

// AdapterProxy : Adapter proxy
type AdapterProxy struct {
	resp            sync.Map
	point           *endpointf.EndpointF
	tarsClient      *transport.TarsClient
	conf            *transport.TarsClientConf
	comm            *Communicator
	obj             *ServantProxy
	failCount       int32
	lastFailCount   int32
	sendCount       int32
	successCount    int32
	status          bool // true for good
	lastSuccessTime int64
	lastBlockTime   int64
	lastCheckTime   int64

	count  int
	closed bool
}

// NewAdapterProxy : Construct an adapter proxy
func NewAdapterProxy(point *endpointf.EndpointF, comm *Communicator) *AdapterProxy {
	c := &AdapterProxy{}
	c.comm = comm
	c.point = point
	proto := "tcp"
	if point.Istcp == 0 {
		proto = "udp"
	}
	conf := &transport.TarsClientConf{
		Proto: proto,
		//NumConnect:   netthread,
		QueueLen:     comm.Client.ClientQueueLen,
		IdleTimeout:  comm.Client.ClientIdleTimeout,
		ReadTimeout:  comm.Client.ClientReadTimeout,
		WriteTimeout: comm.Client.ClientWriteTimeout,
		DialTimeout:  comm.Client.ClientDialTimeout,
	}
	c.conf = conf
	c.tarsClient = transport.NewTarsClient(fmt.Sprintf("%s:%d", point.Host, point.Port), c, conf)
	c.status = true
	return c
}

// ParsePackage : Parse packet from bytes
func (c *AdapterProxy) ParsePackage(buff []byte) (int, int) {
	return c.obj.proto.ParsePackage(buff)
}

// Recv : Recover read channel when closed for timeout
func (c *AdapterProxy) Recv(pkg []byte) {
	defer func() {
		// TODO readCh has a certain probability to be closed after the load, and we need to recover
		// Maybe there is a better way
		if err := recover(); err != nil {
			TLOG.Error("recv pkg painc:", err)
		}
	}()
	packet, err := c.obj.proto.ResponseUnpack(pkg)
	if err != nil {
		TLOG.Error("decode packet error", err.Error())
		return
	}
	if packet.IRequestId == 0 {
		go c.onPush(packet)
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
		//after conf.ReadTimeout, release this goroutine to make sure response package is received by Tars_Invoke().
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
	sbuf, err := c.obj.proto.RequestPack(req)
	if err != nil {
		TLOG.Debug("protocol wrong:", req.IRequestId)
		return err
	}
	return c.tarsClient.Send(sbuf)
}

// GetPoint : Get an endpoint
func (c *AdapterProxy) GetPoint() *endpointf.EndpointF {
	return c.point
}

// Close : Close the client
func (c *AdapterProxy) Close() {
	c.tarsClient.Close()
	c.closed = true
}

func (c *AdapterProxy) sendAdd() {
	atomic.AddInt32(&c.sendCount, 1)
}

func (c *AdapterProxy) succssAdd() {
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
	c.status = true
}

func (c *AdapterProxy) checkActive() (firstTime bool, needCheck bool) {
	if c.closed {
		return false, false
	}

	now := time.Now().Unix()
	if c.status {
		//check if healthy
		if (now-c.lastSuccessTime) >= failInterval && c.lastFailCount >= fainN {
			c.status = false
			c.lastBlockTime = now
			return true, false
		}
		if (now - c.lastCheckTime) >= checkTime {
			if c.failCount >= overN && (float32(c.failCount)/float32(c.sendCount)) >= failRatio {
				c.status = false
				c.lastBlockTime = now
				return true, false
			}
			c.lastCheckTime = now
			return false, false
		}
		return false, false
	}

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
	}
	//TODO: support push msg
}
