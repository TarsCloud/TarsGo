package transport

import (
	"context"
	"errors"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/TarsCloud/TarsGo/tars/util/rtimer"
)

// TarsClientConf is tars client side config
type TarsClientConf struct {
	Proto        string
	ClientProto  ClientProtocol
	QueueLen     int
	IdleTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	DialTimeout  time.Duration
}

// TarsClient is struct for tars client.
type TarsClient struct {
	address string
	//TODO remove it
	conn *connection

	cp        ClientProtocol
	conf      *TarsClientConf
	sendQueue chan []byte
	//recvQueue chan []byte
}

type connection struct {
	tc *TarsClient

	conn     net.Conn
	connLock *sync.Mutex

	isClosed    bool
	idleTime    time.Time
	invokeNum   int32
	dialTimeout time.Duration
}

// NewTarsClient new tars client and init it .
func NewTarsClient(address string, cp ClientProtocol, conf *TarsClientConf) *TarsClient {
	if conf.QueueLen <= 0 {
		conf.QueueLen = 100
	}
	sendQueue := make(chan []byte, conf.QueueLen)
	tc := &TarsClient{conf: conf, address: address, cp: cp, sendQueue: sendQueue}
	tc.conn = &connection{tc: tc, isClosed: true, connLock: &sync.Mutex{}, dialTimeout: conf.DialTimeout}
	return tc
}

// ReConnect established the client connection with the server.
func (tc *TarsClient) ReConnect() error {
	return tc.conn.ReConnect()
}

// Send sends the request to the server as []byte.
func (tc *TarsClient) Send(req []byte) error {
	if err := tc.ReConnect(); err != nil {
		return err
	}

	// avoid full sendQueue that cause sending block
	var timerC <-chan struct{}
	if tc.conf.WriteTimeout > 0 {
		timerC = rtimer.After(tc.conf.WriteTimeout)
	}

	select {
	case <-timerC:
		return errors.New("tars client write timeout")
	case tc.sendQueue <- req:
	}

	return nil
}

// Close close the client connection with the server.
func (tc *TarsClient) Close() {
	w := tc.conn
	if !w.isClosed && w.conn != nil {
		w.isClosed = true
		w.conn.Close()
	}
}

func (c *connection) send(conn net.Conn, connDone chan bool) {
	var req []byte
	t := time.NewTicker(time.Second)
	defer t.Stop()
	for {
		select {
		case <-connDone: // connection closed
			return
		default:
			select {
			case req = <-c.tc.sendQueue: // Fetch jobs
			case <-t.C:
				if c.isClosed {
					return
				}
				// TODO: check one-way invoke for idle detect
				if c.invokeNum == 0 && c.idleTime.Add(c.tc.conf.IdleTimeout).Before(time.Now()) {
					c.close(conn)
					return
				}
				continue
			}
		}
		atomic.AddInt32(&c.invokeNum, 1)
		if c.tc.conf.WriteTimeout != 0 {
			conn.SetWriteDeadline(time.Now().Add(c.tc.conf.WriteTimeout))
		}
		c.idleTime = time.Now()
		_, err := conn.Write(req)
		if err != nil {
			//TODO add retry time
			c.tc.sendQueue <- req
			TLOG.Error("send request error:", err)
			c.close(conn)
			return
		}
	}
}

func (c *connection) recv(conn net.Conn, connDone chan bool) {
	defer func() {
		connDone <- true
	}()
	buffer := make([]byte, 1024*4)
	var currBuffer []byte
	var n int
	var err error
	for {
		if c.tc.conf.ReadTimeout != 0 {
			conn.SetReadDeadline(time.Now().Add(c.tc.conf.ReadTimeout))
		}
		n, err = conn.Read(buffer)
		if err != nil {
			netErr, ok := err.(net.Error)
			if ok && netErr.Timeout() && netErr.Temporary() {
				continue // no data, not error
			}
			if _, ok := err.(*net.OpError); ok {
				TLOG.Errorf("net.OpError: %v, error: %v", conn.RemoteAddr(), err)
				c.close(conn)
				return // connection is closed
			}
			if err == io.EOF {
				TLOG.Debugf("connection closed by remote: %v, error: %v", conn.RemoteAddr(), err)
			} else {
				TLOG.Error("read package error:", err)
			}
			c.close(conn)
			return
		}
		currBuffer = append(currBuffer, buffer[:n]...)
		for {
			pkgLen, status := c.tc.cp.ParsePackage(currBuffer)
			if status == PACKAGE_LESS {
				break
			}
			if status == PACKAGE_FULL {
				atomic.AddInt32(&c.invokeNum, -1)
				pkg := make([]byte, pkgLen)
				copy(pkg, currBuffer[0:pkgLen])
				currBuffer = currBuffer[pkgLen:]
				go c.tc.cp.Recv(pkg)
				if len(currBuffer) > 0 {
					continue
				}
				currBuffer = nil
				break
			}
			TLOG.Error("parse package error")
			c.close(conn)
			return
		}
	}
}

func (c *connection) ReConnect() (err error) {
	c.connLock.Lock()
	if c.isClosed {
		TLOG.Debug("Connect:", c.tc.address)
		c.conn, err = net.DialTimeout(c.tc.conf.Proto, c.tc.address, c.dialTimeout)

		if err != nil {
			c.connLock.Unlock()
			return err
		}
		if c.tc.conf.Proto == "tcp" {
			if c.conn != nil {
				c.conn.(*net.TCPConn).SetKeepAlive(true)
			}
		}
		c.idleTime = time.Now()
		c.isClosed = false
		connDone := make(chan bool, 1)
		go c.recv(c.conn, connDone)
		go c.send(c.conn, connDone)
	}
	c.connLock.Unlock()
	return nil
}

func (c *connection) close(conn net.Conn) {
	c.connLock.Lock()
	c.isClosed = true
	if conn != nil {
		conn.Close()
	}
	c.connLock.Unlock()
}

// GraceClose close client gracefully
func (c *TarsClient) GraceClose(ctx context.Context) {
	tk := time.NewTicker(time.Millisecond * 500)
	defer tk.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-tk.C:
			TLOG.Debugf("wait grace invoke %d", c.conn.invokeNum)
			if atomic.LoadInt32(&c.conn.invokeNum) < 0 {
				c.Close()
				return
			}
		}
	}
}
