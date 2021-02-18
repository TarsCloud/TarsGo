package transport

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/TarsCloud/TarsGo/tars/util/rogger"
)

// TLOG  is logger for transport.
var TLOG = rogger.GetLogger("TLOG")

// ServerHandler  is interface with listen and handler method
type ServerHandler interface {
	Listen() error
	Handle() error
	OnShutdown()
	CloseIdles(n int64) bool
}

// TarsServerConf server config for tars server side.
type TarsServerConf struct {
	Proto          string
	Address        string
	MaxInvoke      int32
	AcceptTimeout  time.Duration
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	HandleTimeout  time.Duration
	IdleTimeout    time.Duration
	QueueCap       int
	TCPReadBuffer  int
	TCPWriteBuffer int
	TCPNoDelay     bool
}

// TarsServer tars server struct.
type TarsServer struct {
	svr        ServerProtocol
	conf       *TarsServerConf
	handle     ServerHandler
	lastInvoke time.Time
	idleTime   time.Time
	isClosed   int32
	numInvoke  int32
	numConn    int32
}

// NewTarsServer new TarsServer and init with conf.
func NewTarsServer(svr ServerProtocol, conf *TarsServerConf) *TarsServer {
	ts := &TarsServer{svr: svr, conf: conf}
	ts.isClosed = 0
	ts.lastInvoke = time.Now()
	return ts
}

func (ts *TarsServer) getHandler() (sh ServerHandler) {
	if ts.conf.Proto == "tcp" {
		sh = &tcpHandler{conf: ts.conf, ts: ts}
	} else if ts.conf.Proto == "udp" {
		sh = &udpHandler{conf: ts.conf, ts: ts}
	} else {
		panic("unsupport protocol: " + ts.conf.Proto)
	}
	return
}

// Serve accepts incoming connections
func (ts *TarsServer) Serve() error {
	if ts.handle == nil {
		panic("handle is nil")
	}
	return ts.handle.Handle()
}

// Listen listens on the network address
func (ts *TarsServer) Listen() error {
	ts.handle = ts.getHandler()
	return ts.handle.Listen()
}

// Shutdown try to shutdown server gracefully.
func (ts *TarsServer) Shutdown(ctx context.Context) error {
	// step 1: close listeners, notify client reconnect
	atomic.StoreInt32(&ts.isClosed, 1)
	ts.handle.OnShutdown()

	// step 2: wait and close idle connections
	watchInterval := time.Millisecond * 500
	tk := time.NewTicker(watchInterval)
	defer tk.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-tk.C:
			if ts.handle.CloseIdles(2) {
				return nil
			}
		}
	}
}

// GetConfig gets the tars server config.
func (ts *TarsServer) GetConfig() *TarsServerConf {
	return ts.conf
}

// IsZombie show whether the server is hanged by the request.
func (ts *TarsServer) IsZombie(timeout time.Duration) bool {
	conf := ts.GetConfig()
	return conf.MaxInvoke != 0 && ts.numInvoke == conf.MaxInvoke && ts.lastInvoke.Add(timeout).Before(time.Now())
}

func (ts *TarsServer) invoke(ctx context.Context, pkg []byte) []byte {
	cfg := ts.conf
	var rsp []byte
	if cfg.HandleTimeout == 0 {
		rsp = ts.svr.Invoke(ctx, pkg)
	} else {
		invokeDone, cancelFunc := context.WithTimeout(context.Background(), cfg.HandleTimeout)
		go func() {
			rsp = ts.svr.Invoke(ctx, pkg)
			cancelFunc()
		}()
		select {
		case <-invokeDone.Done():
			if len(rsp) == 0 { // The rsp must be none-empty
				rsp = ts.svr.InvokeTimeout(pkg)
			}
		}
	}
	return rsp
}
