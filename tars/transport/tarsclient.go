package transport

import (
	"context"
	"crypto/tls"
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
	TlsConfig    *tls.Config
}

// TarsClient is struct for tars client.
type TarsClient struct {
	address string
	// TODO remove it
	conn *connection

	protocol      ClientProtocol
	config        *TarsClientConf
	sendQueue     chan sendMsg
	sendFailQueue chan sendMsg
}

type sendMsg struct {
	req   []byte
	retry uint8
}

type connection struct {
	client *TarsClient

	conn     net.Conn
	connLock sync.Mutex

	isClosed    bool
	idleTime    time.Time
	invokeNum   int32
	dialTimeout time.Duration
}

// NewTarsClient new tars client and init it .
func NewTarsClient(address string, protocol ClientProtocol, config *TarsClientConf) *TarsClient {
	if config.QueueLen <= 0 {
		config.QueueLen = 100
	}
	client := &TarsClient{
		config:        config,
		address:       address,
		protocol:      protocol,
		sendQueue:     make(chan sendMsg, config.QueueLen),
		sendFailQueue: make(chan sendMsg, 1),
	}
	client.conn = &connection{client: client, isClosed: true, dialTimeout: config.DialTimeout}
	return client
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
	if tc.config.WriteTimeout > 0 {
		timerC = rtimer.After(tc.config.WriteTimeout)
	}

	select {
	case <-timerC:
		return errors.New("tars client write timeout")
	case tc.sendQueue <- sendMsg{req: req}:
		return nil
	}
}

// Close the client connection with the server.
func (tc *TarsClient) Close() {
	w := tc.conn
	if !w.isClosed && w.conn != nil {
		w.isClosed = true
		_ = w.conn.Close()
	}
}

// GraceClose close client gracefully
func (tc *TarsClient) GraceClose(ctx context.Context) {
	tk := time.NewTicker(time.Millisecond * 500)
	defer tk.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-tk.C:
			TLOG.Debugf("wait grace invoke %d", tc.conn.invokeNum)
			if atomic.LoadInt32(&tc.conn.invokeNum) <= 0 {
				tc.Close()
				return
			}
		}
	}
}

func (c *connection) ReConnect() (err error) {
	c.connLock.Lock()
	defer c.connLock.Unlock()
	if c.isClosed {
		TLOG.Debug("Connect:", c.client.address, "Proto:", c.client.config.Proto)
		if c.client.config.Proto == "ssl" {
			dialer := &net.Dialer{Timeout: c.dialTimeout}
			c.conn, err = tls.DialWithDialer(dialer, "tcp", c.client.address, c.client.config.TlsConfig)
		} else {
			c.conn, err = net.DialTimeout(c.client.config.Proto, c.client.address, c.dialTimeout)
		}

		if err != nil {
			go c.client.protocol.OnClose(c.client.address)
			return err
		}
		if c.client.config.Proto == "tcp" {
			if c.conn != nil {
				c.conn.(*net.TCPConn).SetKeepAlive(true)
			}
		}
		c.idleTime = time.Now()
		c.isClosed = false
		go c.client.protocol.OnConnect(c.client.address)
		connDone := make(chan bool, 1)
		go c.recv(c.conn, connDone)
		go c.send(c.conn, connDone)
	}
	return nil
}

func (c *connection) send(conn net.Conn, connDone chan bool) {
	var m sendMsg
	t := time.NewTicker(time.Second)
	defer t.Stop()
	for {
		select {
		case <-connDone: // connection closed
			return
		default:
		}
		// get sendMsg
		select {
		case m = <-c.client.sendFailQueue: // Send failure queue messages first
		default:
			select {
			case m = <-c.client.sendQueue: // Fetch jobs
			case <-t.C:
				if c.isClosed {
					return
				}
				// TODO: check one-way invoke for idle detect
				if c.invokeNum == 0 && c.idleTime.Add(c.client.config.IdleTimeout).Before(time.Now()) {
					c.close(conn)
					return
				}
				continue
			}
		}
		atomic.AddInt32(&c.invokeNum, 1)
		if c.client.config.WriteTimeout != 0 {
			if err := conn.SetWriteDeadline(time.Now().Add(c.client.config.WriteTimeout)); err != nil {
				TLOG.Errorf("set write deadline error: %v", err)
			}
		}
		c.idleTime = time.Now()
		_, err := conn.Write(m.req)
		if err != nil {
			// TODO add retry times
			m.retry++
			TLOG.Errorf("send request retry: %d, error: %v", m.retry, err)
			c.client.sendFailQueue <- m
			c.close(conn)
			if err != net.ErrClosed {
				return
			}

			// connection closed, try to reconnect once
			if err = c.ReConnect(); err != nil {
				TLOG.Errorf("send request reconnect error: %v", err)
			}
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
		if c.client.config.ReadTimeout != 0 {
			if err = conn.SetReadDeadline(time.Now().Add(c.client.config.ReadTimeout)); err != nil {
				TLOG.Errorf("set read deadline error: %v", err)
			}
		}
		n, err = conn.Read(buffer)
		if err != nil {
			if isNoDataError(err) {
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
				TLOG.Errorf("read package error: %v", err)
			}
			c.close(conn)
			return
		}
		currBuffer = append(currBuffer, buffer[:n]...)
		for {
			pkgLen, status := c.client.protocol.ParsePackage(currBuffer)
			if status == PackageLess {
				break
			}
			if status == PackageFull {
				atomic.AddInt32(&c.invokeNum, -1)
				pkg := make([]byte, pkgLen)
				copy(pkg, currBuffer[0:pkgLen])
				currBuffer = currBuffer[pkgLen:]
				go c.client.protocol.Recv(pkg)
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

func (c *connection) close(conn net.Conn) {
	c.connLock.Lock()
	defer c.connLock.Unlock()
	c.isClosed = true
	if conn != nil {
		_ = conn.Close()
	}
	go c.client.protocol.OnClose(c.client.address)
}
