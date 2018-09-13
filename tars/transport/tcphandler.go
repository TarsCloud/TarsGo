package transport

import (
	"io"
	"net"
	"reflect"
	"sync/atomic"
	"time"
)

type tcpHandler struct {
	conf *TarsServerConf

	lis *net.TCPListener
	ts  *TarsServer

	acceptNum   int32
	invokeNum   int32
	readBuffer  int
	writeBuffer int
	tcpNoDelay  bool
	jobs        chan *connectionHandler
	idleTime    time.Time
}
type connectionHandler struct {
	conn *net.TCPConn
	pkg  []byte
}

func (h *tcpHandler) Listen() (err error) {
	cfg := h.conf
	addr, err := net.ResolveTCPAddr("tcp4", cfg.Address)
	if err != nil {
		return err
	}
	h.lis, err = net.ListenTCP("tcp4", addr)
	TLOG.Info("Listening on", cfg.Address)
	return
}

func (h *tcpHandler) Handle() error {
	cfg := h.conf
	if cfg.MaxInvoke > 0 {
		for i := 0; i < int(cfg.MaxInvoke); i++ {
			go h.worker()
		}
	}
	for !h.ts.isClosed {
		h.lis.SetDeadline(time.Now().Add(cfg.AcceptTimeout)) // set accept timeout
		conn, err := h.lis.AcceptTCP()
		if err != nil {
			if !isNoDataError(err) {
				TLOG.Errorf("Accept error: %v", err)
			} else if conn != nil {
				conn.SetKeepAlive(true)
			}
			continue
		}
		go func(conn *net.TCPConn) {
			TLOG.Debug("TCP accept:", conn.RemoteAddr())
			atomic.AddInt32(&h.acceptNum, 1)
			conn.SetReadBuffer(cfg.TCPReadBuffer)
			conn.SetWriteBuffer(cfg.TCPWriteBuffer)
			conn.SetNoDelay(cfg.TCPNoDelay)
			h.recv(conn)
			atomic.AddInt32(&h.acceptNum, -1)
		}(conn)
	}
	return nil
}

func (h *tcpHandler) worker() {
	for {
		select {
		case ch := <-h.jobs:
			rsp := h.ts.invoke(ch.pkg)
			if _, err := ch.conn.Write(rsp); err != nil {
				TLOG.Errorf("send pkg to %v failed %v", ch.conn.RemoteAddr(), err)
			}
		}
	}

}
func (h *tcpHandler) recv(conn *net.TCPConn) {
	defer conn.Close()
	cfg := h.conf
	buffer := make([]byte, 1024*4)
	var currBuffer []byte // need a deep copy of buffer
	h.idleTime = time.Now()
	var n int
	var err error
	for !h.ts.isClosed {
		if cfg.ReadTimeout != 0 {
			conn.SetReadDeadline(time.Now().Add(cfg.ReadTimeout))
		}
		n, err = conn.Read(buffer)
		if err != nil {
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
				TLOG.Error("read packge error:", reflect.TypeOf(err), err)
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
				if h.conf.MaxInvoke > 0 {
					ch := &connectionHandler{conn: conn, pkg: pkg[:]}
					h.jobs <- ch
				} else {
					go func(pkg []byte) {
						rsp := h.ts.invoke(pkg)
						if _, err := conn.Write(rsp); err != nil {
							TLOG.Errorf("send pkg to %v failed %v", conn.RemoteAddr(), err)
						}
					}(pkg[:])
				}
				if len(currBuffer) > 0 {
					continue
				}
				currBuffer = nil
				break
			}
			TLOG.Errorf("parse packge error %s %v", conn.RemoteAddr(), err)
			return
		}
	}
}
