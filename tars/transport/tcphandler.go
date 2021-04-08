package transport

import (
	"context"
	"fmt"
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
)

type tcpHandler struct {
	conf *TarsServerConf

	lis *net.TCPListener
	ts  *TarsServer

	readBuffer  int
	writeBuffer int
	tcpNoDelay  bool
	gpool       *gpool.Pool

	conns sync.Map

	isListenClosed int32
}

type connInfo struct {
	conn      *net.TCPConn
	idleTime  int64
	numInvoke int32
	writeLock sync.Mutex
}

func (h *tcpHandler) Listen() (err error) {
	cfg := h.conf
	ln, err := grace.CreateListener("tcp", cfg.Address)
	if err == nil {
		TLOG.Infof("Listening on %s", cfg.Address)
		h.lis = ln.(*net.TCPListener)
	} else {
		TLOG.Infof("Listening on %s error: %v", cfg.Address, err)
	}

	// init goroutine pool
	if cfg.MaxInvoke > 0 {
		h.gpool = gpool.NewPool(int(cfg.MaxInvoke), cfg.QueueCap)
	}

	return err
}

func (h *tcpHandler) getConnContext(connSt *connInfo) context.Context {
	ctx := current.ContextWithTarsCurrent(context.Background())
	ipPort := strings.Split(connSt.conn.RemoteAddr().String(), ":")
	current.SetClientIPWithContext(ctx, ipPort[0])
	current.SetClientPortWithContext(ctx, ipPort[1])
	current.SetRecvPkgTsFromContext(ctx, time.Now().UnixNano()/1e6)

	return ctx
}

func (h *tcpHandler) handleConn(connSt *connInfo, pkg []byte) {
	handler := func() {
		defer atomic.AddInt32(&connSt.numInvoke, -1)

		ctx := h.getConnContext(connSt)
		rsp := h.ts.invoke(ctx, pkg)

		cPacketType, ok := current.GetPacketTypeFromContext(ctx)
		if !ok {
			TLOG.Error("Failed to GetPacketTypeFromContext")
		}
		if cPacketType == basef.TARSONEWAY {
			return
		}

		connSt.writeLock.Lock()
		if _, err := connSt.conn.Write(rsp); err != nil {
			TLOG.Errorf("send pkg to %v failed %v", connSt.conn.RemoteAddr(), err)
		}
		connSt.writeLock.Unlock()
	}

	cfg := h.conf
	if cfg.MaxInvoke > 0 { // use goroutine pool
		h.gpool.JobQueue <- handler
	} else {
		go handler()
	}
}

func (h *tcpHandler) Handle() error {
	cfg := h.conf
	for {
		if atomic.LoadInt32(&h.ts.isClosed) == 1 {
			TLOG.Errorf("Close accept %s %d", h.conf.Address, os.Getpid())
			atomic.StoreInt32(&h.isListenClosed, 1)
			break
		}
		if cfg.AcceptTimeout > 0 {
			// set accept timeout
			h.lis.SetDeadline(time.Now().Add(cfg.AcceptTimeout))
		}
		conn, err := h.lis.AcceptTCP()
		if err != nil {
			if !isNoDataError(err) {
				TLOG.Errorf("Accept error: %v", err)
			} else if conn != nil {
				conn.SetKeepAlive(true)
			}
			continue
		}
		atomic.AddInt32(&h.ts.numConn, 1)
		go func(conn *net.TCPConn) {
			fd, _ := conn.File()
			defer fd.Close()
			key := fmt.Sprintf("%v", fd.Fd())
			TLOG.Debugf("TCP accept: %s, %d, fd: %s", conn.RemoteAddr(), os.Getpid(), key)
			conn.SetReadBuffer(cfg.TCPReadBuffer)
			conn.SetWriteBuffer(cfg.TCPWriteBuffer)
			conn.SetNoDelay(cfg.TCPNoDelay)

			cf := &connInfo{conn: conn}
			h.conns.Store(key, cf)
			h.recv(cf)
			h.conns.Delete(key)
		}(conn)
	}
	if h.gpool != nil {
		h.gpool.Release()
	}
	return nil
}

func (h *tcpHandler) OnShutdown() {
	// close listeners
	h.lis.SetDeadline(time.Now())
	if atomic.LoadInt32(&h.isListenClosed) == 1 {
		h.sendCloseMsg()
		atomic.StoreInt32(&h.isListenClosed, 2)
	}
}

func (h *tcpHandler) sendCloseMsg() {
	// send close-package
	closeMsg := h.ts.svr.GetCloseMsg()
	h.conns.Range(func(key, val interface{}) bool {
		conn := val.(*connInfo)
		conn.conn.SetReadDeadline(time.Now())
		// send a reconnect-message
		TLOG.Debugf("send close message to %v", conn.conn.RemoteAddr())
		conn.writeLock.Lock()
		conn.conn.Write(closeMsg)
		conn.writeLock.Unlock()
		return true
	})
}

//CloseIdles close all idle connections(no active package within n secnods)
func (h *tcpHandler) CloseIdles(n int64) bool {
	if atomic.LoadInt32(&h.isListenClosed) == 0 {
		// hack: create new connection to avoid acceptTCP hanging
		TLOG.Debugf("Hack msg to %s", h.conf.Address)
		if conn, err := net.Dial("tcp", h.conf.Address); err == nil {
			conn.Close()
		}
	}
	if atomic.LoadInt32(&h.isListenClosed) == 1 {
		h.sendCloseMsg()
		atomic.StoreInt32(&h.isListenClosed, 2)
	}

	allClosed := true
	h.conns.Range(func(key, val interface{}) bool {
		conn := val.(*connInfo)
		TLOG.Debugf("num invoke %d %v", atomic.LoadInt32(&conn.numInvoke), conn.idleTime+n > time.Now().Unix())
		if atomic.LoadInt32(&conn.numInvoke) > 0 || conn.idleTime+n > time.Now().Unix() {
			allClosed = false
		}
		return true
	})
	return allClosed
}

func (h *tcpHandler) recv(connSt *connInfo) {
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

		ctx := h.getConnContext(connSt)
		h.ts.svr.DoClose(ctx)

		connSt.idleTime = 0
	}()

	cfg := h.conf
	buffer := make([]byte, 1024*4)
	var currBuffer []byte // need a deep copy of buffer
	//TODO: change to gtime
	connSt.idleTime = time.Now().Unix()
	var n int
	var err error
	for {
		if atomic.LoadInt32(&h.ts.isClosed) == 1 {
			// set short deadline to clear connection buffer
			conn.SetReadDeadline(time.Now().Add(time.Millisecond * 100))
		} else if cfg.ReadTimeout > 0 {
			conn.SetReadDeadline(time.Now().Add(cfg.ReadTimeout))
		}
		connSt.idleTime = time.Now().Unix()
		n, err = conn.Read(buffer)
		// TLOG.Debugf("%s closed: %d, read %d, nil buff: %d, err: %v", h.ts.conf.Address, atomic.LoadInt32(&h.ts.isClosed), n, len(currBuffer), err)
		if err != nil {
			if atomic.LoadInt32(&h.ts.isClosed) == 1 && currBuffer == nil {
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
			pkgLen, status := h.ts.svr.ParsePackage(currBuffer)
			if status == PACKAGE_LESS {
				break
			}
			if status == PACKAGE_FULL {
				atomic.AddInt32(&connSt.numInvoke, 1)
				pkg := make([]byte, pkgLen)
				copy(pkg, currBuffer[:pkgLen])
				currBuffer = currBuffer[pkgLen:]
				h.handleConn(connSt, pkg)
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
