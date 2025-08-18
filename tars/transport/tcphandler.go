package transport

import (
	"context"
	"crypto/tls"
	"github.com/google/uuid"
	"io"
	"net"
	"os"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/TarsCloud/TarsGo/tars/protocol/res/basef"
	"github.com/TarsCloud/TarsGo/tars/util/current"
	"github.com/TarsCloud/TarsGo/tars/util/gpool"
	"github.com/TarsCloud/TarsGo/tars/util/grace"
	"github.com/TarsCloud/TarsGo/tars/util/gtime"
)

type tcpHandler struct {
	config *TarsServerConf

	server         *TarsServer
	listener       net.Listener
	tcpListener    *net.TCPListener
	isListenClosed int32

	pool  *gpool.Pool
	conns sync.Map
}

type connInfo struct {
	uuid      string
	conn      net.Conn
	idleTime  int64
	numInvoke int32
}

func (t *tcpHandler) Listen() (err error) {
	uuid.EnableRandPool()
	cfg := t.config
	t.listener, err = grace.CreateListener("tcp", cfg.Address)
	if err != nil {
		TLOG.Errorf("Listening on %s error: %v", cfg.Address, err)
		return err
	}

	TLOG.Infof("Listening on %s", cfg.Address)
	t.tcpListener = t.listener.(*net.TCPListener)
	if t.config.TlsConfig != nil {
		t.listener = tls.NewListener(t.listener, t.config.TlsConfig)
	}

	// init goroutine pool
	if cfg.MaxInvoke > 0 {
		t.pool = gpool.NewPool(int(cfg.MaxInvoke), cfg.QueueCap)
	}

	return nil
}

func (t *tcpHandler) getConnContext(connSt *connInfo) context.Context {
	ctx := current.ContextWithTarsCurrent(context.Background())
	ipPort := strings.Split(connSt.conn.RemoteAddr().String(), ":")
	current.SetUUIDWithContext(ctx, connSt.uuid)
	current.SetClientIPWithContext(ctx, ipPort[0])
	current.SetClientPortWithContext(ctx, ipPort[1])
	current.SetRecvPkgTsFromContext(ctx, time.Now().UnixNano()/1e6)
	current.SetRawConnWithContext(ctx, connSt.conn, nil)
	return ctx
}

func (t *tcpHandler) handleConn(connSt *connInfo, pkg []byte) {
	// recvPkgTs are more accurate
	ctx := t.getConnContext(connSt)
	atomic.AddInt32(&connSt.numInvoke, 1)
	handler := func() {
		defer atomic.AddInt32(&connSt.numInvoke, -1)
		rsp := t.server.invoke(ctx, pkg)

		cPacketType, ok := current.GetPacketTypeFromContext(ctx)
		if !ok {
			TLOG.Error("Failed to GetPacketTypeFromContext")
		}
		if cPacketType == basef.TARSONEWAY {
			return
		}

		if _, err := connSt.conn.Write(rsp); err != nil {
			TLOG.Errorf("send pkg to %v failed %v", connSt.conn.RemoteAddr(), err)
		}
	}

	cfg := t.config
	if cfg.MaxInvoke > 0 { // use goroutine pool
		t.pool.JobQueue <- handler
	} else {
		go handler()
	}
}

func (t *tcpHandler) Handle() error {
	cfg := t.config
	for {
		if atomic.LoadInt32(&t.server.isClosed) == 1 {
			TLOG.Errorf("Close accept %s %d", t.config.Address, os.Getpid())
			atomic.StoreInt32(&t.isListenClosed, 1)
			break
		}
		if cfg.AcceptTimeout > 0 {
			// set accept timeout
			if err := t.tcpListener.SetDeadline(time.Now().Add(cfg.AcceptTimeout)); err != nil {
				TLOG.Errorf("SetDeadline error: %v", err)
			}
		}
		conn, err := t.listener.Accept()
		if err != nil {
			if !isNoDataError(err) {
				TLOG.Errorf("Accept error: %v", err)
			} else if conn != nil {
				if c, ok := conn.(*net.TCPConn); ok {
					if err = c.SetKeepAlive(true); err != nil {
						TLOG.Errorf("SetKeepAlive error: %v", err)
					}
				}
			}
			continue
		}
		atomic.AddInt32(&t.server.numConn, 1)
		go func(conn net.Conn) {
			key := conn.RemoteAddr().String()
			switch c := conn.(type) {
			case *net.TCPConn:
				TLOG.Debugf("TCP accept: %s, %d", conn.RemoteAddr(), os.Getpid())
				c.SetReadBuffer(cfg.TCPReadBuffer)
				c.SetWriteBuffer(cfg.TCPWriteBuffer)
				c.SetNoDelay(cfg.TCPNoDelay)
			case *tls.Conn:
				TLOG.Debugf("TLS accept: %s, %d", conn.RemoteAddr(), os.Getpid())
			}
			cf := &connInfo{conn: conn, uuid: uuid.New().String()}
			t.conns.Store(key, cf)
			t.recv(cf)
			t.conns.Delete(key)
		}(conn)
	}
	if t.pool != nil {
		t.pool.Release()
	}
	return nil
}

func (t *tcpHandler) OnShutdown() {
	// close listeners
	t.tcpListener.SetDeadline(time.Now())
	if atomic.LoadInt32(&t.isListenClosed) == 1 {
		t.sendCloseMsg()
		atomic.StoreInt32(&t.isListenClosed, 2)
	}
}

func (t *tcpHandler) sendCloseMsg() {
	// send close-package
	closeMsg := t.server.protocol.GetCloseMsg()
	t.conns.Range(func(key, val interface{}) bool {
		conn := val.(*connInfo)
		if err := conn.conn.SetReadDeadline(time.Now()); err != nil {
			TLOG.Errorf("SetReadDeadline: %w", err)
		}
		// send a reconnect-message
		TLOG.Debugf("send close message to %v", conn.conn.RemoteAddr())
		if _, err := conn.conn.Write(closeMsg); err != nil {
			TLOG.Errorf("send closeMsg to %v failed %v", conn.conn.RemoteAddr(), err)
		}
		return true
	})
}

// CloseIdles close all idle connections(no active package within n secnods)
func (t *tcpHandler) CloseIdles(n int64) bool {
	if atomic.LoadInt32(&t.isListenClosed) == 0 {
		// hack: create new connection to avoid acceptTCP hanging
		TLOG.Debugf("Hack msg to %s", t.config.Address)
		if conn, err := net.Dial("tcp", t.config.Address); err == nil {
			conn.Close()
		}
	}
	if atomic.LoadInt32(&t.isListenClosed) == 1 {
		t.sendCloseMsg()
		atomic.StoreInt32(&t.isListenClosed, 2)
	}

	allClosed := true
	t.conns.Range(func(key, val interface{}) bool {
		conn := val.(*connInfo)
		TLOG.Debugf("num invoke %d %v", atomic.LoadInt32(&conn.numInvoke), conn.idleTime+n > time.Now().Unix())
		if atomic.LoadInt32(&conn.numInvoke) > 0 || conn.idleTime+n > time.Now().Unix() {
			allClosed = false
			return true
		}
		conn.conn.Close()
		return true
	})
	return allClosed
}

func (t *tcpHandler) recv(connSt *connInfo) {
	conn := connSt.conn
	defer func() {
		watchInterval := time.Millisecond * 500
		tk := time.NewTicker(watchInterval)
		defer tk.Stop()
		for range tk.C {
			if atomic.LoadInt32(&connSt.numInvoke) == 0 {
				break
			}
		}
		TLOG.Debugf("Close connection: %v", conn.RemoteAddr())
		conn.Close()

		ctx := t.getConnContext(connSt)
		t.server.protocol.DoClose(ctx)

		connSt.idleTime = 0
	}()

	cfg := t.config
	buffer := make([]byte, 1024*4)
	var currBuffer []byte // need a deep copy of buffer
	connSt.idleTime = gtime.CurrUnixTime
	var n int
	var err error
	for {
		if atomic.LoadInt32(&t.server.isClosed) == 1 {
			// set short deadline to clear connection buffer
			conn.SetReadDeadline(time.Now().Add(time.Millisecond * 100))
		} else if cfg.ReadTimeout > 0 {
			conn.SetReadDeadline(time.Now().Add(cfg.ReadTimeout))
		}
		connSt.idleTime = time.Now().Unix()
		n, err = conn.Read(buffer)
		if err != nil {
			TLOG.Debugf("%s closed: %d, read %d, nil buff: %d, err: %v", t.server.config.Address, atomic.LoadInt32(&t.server.isClosed), n, len(currBuffer), err)
			if atomic.LoadInt32(&t.server.isClosed) == 1 && currBuffer == nil {
				return
			}
			if len(currBuffer) == 0 && connSt.numInvoke == 0 && (connSt.idleTime+int64(cfg.IdleTimeout)/int64(time.Second)) < time.Now().Unix() {
				return
			}
			if isNoDataError(err) {
				continue
			}
			if err == io.EOF {
				TLOG.Debug("connection closed by remote:", conn.RemoteAddr())
			} else {
				TLOG.Error("read package error:", reflect.TypeOf(err), err)
			}
			return
		}
		currBuffer = append(currBuffer, buffer[:n]...)
		for {
			pkgLen, status := t.server.protocol.ParsePackage(currBuffer)
			if status == PackageLess {
				break
			}
			if status == PackageFull {
				pkg := make([]byte, pkgLen)
				copy(pkg, currBuffer[:pkgLen])
				currBuffer = currBuffer[pkgLen:]
				t.handleConn(connSt, pkg)
				if len(currBuffer) > 0 {
					continue
				}
				currBuffer = nil
				break
			}
			TLOG.Errorf("parse package error %s %v", conn.RemoteAddr(), err)
			return
		}
	}
}
