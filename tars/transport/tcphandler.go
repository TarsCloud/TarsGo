package transport

import (
	"context"
	"io"
	"net"
	"reflect"
	"strings"
	"sync/atomic"
	"time"

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
	idleTime    time.Time
	gpool       *gpool.Pool
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
	return err
}

func (h *tcpHandler) handleConn(conn *net.TCPConn, pkg []byte) {
	handler := func() {
		ctx := context.Background()
		remoteAddr := conn.RemoteAddr().String()
		ipPort := strings.Split(remoteAddr, ":")
		ctx = current.ContextWithTarsCurrent(ctx)
		ok := current.SetClientIPWithContext(ctx, ipPort[0])
		if !ok {
			TLOG.Error("Failed to set context with client ip")
		}
		ok = current.SetClientPortWithContext(ctx, ipPort[1])
		if !ok {
			TLOG.Error("Failed to set context with client port")
		}
		rsp := h.ts.invoke(ctx, pkg)
		if _, err := conn.Write(rsp); err != nil {
			TLOG.Errorf("send pkg to %v failed %v", remoteAddr, err)
		}
	}

	cfg := h.conf
	if cfg.MaxInvoke > 0 { // use goroutine pool
		if h.gpool == nil {
			h.gpool = gpool.NewPool(int(cfg.MaxInvoke), cfg.QueueCap)
		}

		h.gpool.JobQueue <- handler
	} else {
		go handler()
	}
}

func (h *tcpHandler) Handle() error {
	cfg := h.conf
	for {
		if atomic.LoadInt32(&h.ts.isClosed) == 1 {
			return nil
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
			TLOG.Debug("TCP accept:", conn.RemoteAddr())
			conn.SetReadBuffer(cfg.TCPReadBuffer)
			conn.SetWriteBuffer(cfg.TCPWriteBuffer)
			conn.SetNoDelay(cfg.TCPNoDelay)
			h.recv(conn)
			atomic.AddInt32(&h.ts.numConn, -1)
		}(conn)
	}
	if h.gpool != nil {
		h.gpool.Release()
	}
	return nil
}

func (h *tcpHandler) recv(conn *net.TCPConn) {
	defer conn.Close()
	cfg := h.conf
	buffer := make([]byte, 1024*4)
	var currBuffer []byte // need a deep copy of buffer
	h.idleTime = time.Now()
	var n int
	var err error
	for {
		if atomic.LoadInt32(&h.ts.isClosed) == 1 {
			// set short deadline to clear connection buffer
			conn.SetReadDeadline(time.Now().Add(time.Millisecond * 10))
		} else if cfg.ReadTimeout > 0 {
			conn.SetReadDeadline(time.Now().Add(cfg.ReadTimeout))
		}
		n, err = conn.Read(buffer)
		if err != nil {
			if atomic.LoadInt32(&h.ts.isClosed) == 1 && len(currBuffer) == 0 {
				return
			}
			if len(currBuffer) == 0 && h.ts.numInvoke == 0 && h.idleTime.Add(cfg.IdleTimeout).Before(time.Now()) {
				return
			}
			h.idleTime = time.Now()
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
				pkg := make([]byte, pkgLen-4)
				copy(pkg, currBuffer[4:pkgLen])
				currBuffer = currBuffer[pkgLen:]
				h.handleConn(conn, pkg)
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
