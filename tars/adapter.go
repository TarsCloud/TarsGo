package tars

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/TarsCloud/TarsGo/tars/protocol/codec"
	"github.com/TarsCloud/TarsGo/tars/protocol/res/endpointf"
	"github.com/TarsCloud/TarsGo/tars/protocol/res/requestf"
	"github.com/TarsCloud/TarsGo/tars/transport"
	"sync"
	"sync/atomic"
	"time"
)

type AdapterProxy struct {
	resp       sync.Map
	point      *endpointf.EndpointF
	tarsClient *transport.TarsClient
	comm       *Communicator
	failCount  int32
	sendCount  int32
	status     bool
	closed     bool
}

func (c *AdapterProxy) New(point *endpointf.EndpointF, comm *Communicator) error {
	c.comm = comm
	c.point = point
	proto := "tcp"
	if point.Istcp == 0 {
		proto = "udp"
	}

	conf := &transport.TarsClientConf{
		Proto: proto,
		//NumConnect:   netthread,
		QueueLen:     ClientQueueLen,
		IdleTimeout:  ClientIdleTimeout,
		ReadTimeout:  ClientReadTimeout,
		WriteTimeout: ClientWriteTimeout,
	}
	c.tarsClient = transport.NewTarsClient(fmt.Sprintf("%s:%d", point.Host, point.Port), c, conf)
	c.status = true
	go c.checkActive()
	return nil
}

func (c *AdapterProxy) ParsePackage(buff []byte) (int, int) {
	return TarsRequest(buff)
}

func (c *AdapterProxy) Recv(pkg []byte) {
	defer func() {
		//TODO readCh在load之后一定几率被超时关闭了,这个时候需要recover恢复
		//或许有更好的办法吧
		if err := recover(); err != nil {
			TLOG.Error("recv pkg painc:", err)
		}
	}()
	packet := requestf.ResponsePacket{}
	err := packet.ReadFrom(codec.NewReader(pkg))
	if err != nil {
		TLOG.Error("decode packet error", err.Error())
		return
	}
	chIF, ok := c.resp.Load(packet.IRequestId)
	if ok {
		ch := chIF.(chan *requestf.ResponsePacket)
		TLOG.Debug("IN:", packet)
		ch <- &packet
	} else {
		TLOG.Error("timeout resp,drop it:", packet.IRequestId)
	}
}

func (c *AdapterProxy) Send(req *requestf.RequestPacket) error {
	TLOG.Debug("send req:", req.IRequestId)
	c.sendAdd()
	sbuf := bytes.NewBuffer(nil)
	sbuf.Write(make([]byte, 4))
	os := codec.NewBuffer()
	req.WriteTo(os)
	bs := os.ToBytes()
	sbuf.Write(bs)
	len := sbuf.Len()
	binary.BigEndian.PutUint32(sbuf.Bytes(), uint32(len))
	return c.tarsClient.Send(sbuf.Bytes())
}

func (c *AdapterProxy) GetPoint() *endpointf.EndpointF {
	return c.point
}

func (c *AdapterProxy) Close() {
	c.tarsClient.Close()
	c.closed = true
}

func (a *AdapterProxy) sendAdd() {
	atomic.AddInt32(&a.sendCount, 1)
}
func (a *AdapterProxy) failAdd() {
	atomic.AddInt32(&a.failCount, 1)
}
func (a *AdapterProxy) reset() {
	atomic.SwapInt32(&a.sendCount, 0)
	atomic.SwapInt32(&a.failCount, 0)
}
func (a *AdapterProxy) checkActive() {
	loop := time.NewTicker(AdapterProxyTicker)
	count := 0 //每分钟探测一次死掉的节点是否恢复
	for range loop.C {
		if a.closed {
			loop.Stop()
			return
		}
		if a.failCount > a.sendCount/2 {
			a.status = false
		}
		if !a.status && count > AdapterProxyResetCount {
			//TODO USE TAFPING INSTEAD
			a.reset()
			a.status = true
			count = 0
		}
		count++
	}
}
