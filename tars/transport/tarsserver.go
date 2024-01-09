package transport

import (
	"context"
	"crypto/tls"
	"sync/atomic"
	"time"

	"github.com/TarsCloud/TarsGo/tars/util/rogger"
)

// TLOG  is logger for transport.
var TLOG = rogger.GetLogger("TLOG")

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
	TlsConfig      *tls.Config
}

// TarsServer tars server struct.
type TarsServer struct {
	protocol   ServerProtocol
	config     *TarsServerConf
	handle     ServerHandler
	lastInvoke time.Time
	isClosed   int32
	numInvoke  int32
	numConn    int32
}

// NewTarsServer new TarsServer and init with config.
func NewTarsServer(protocol ServerProtocol, config *TarsServerConf) *TarsServer {
	ts := &TarsServer{protocol: protocol, config: config}
	ts.isClosed = 0
	ts.lastInvoke = time.Now()
	return ts
}

func (ts *TarsServer) getHandler() (sh ServerHandler) {
	if ts.config.Proto == "tcp" {
		sh = &tcpHandler{config: ts.config, server: ts}
	} else if ts.config.Proto == "udp" {
		sh = &udpHandler{config: ts.config, server: ts}
	} else {
		panic("unsupport protocol: " + ts.config.Proto)
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
	return ts.config
}

// IsZombie show whether the server is hanged by the request.
func (ts *TarsServer) IsZombie(timeout time.Duration) bool {
	conf := ts.GetConfig()
	return conf.MaxInvoke != 0 && ts.numInvoke == conf.MaxInvoke && ts.lastInvoke.Add(timeout).Before(time.Now())
}

func (ts *TarsServer) invoke(ctx context.Context, pkg []byte) []byte {
	cfg := ts.config
	var rsp []byte
	if cfg.HandleTimeout == 0 {
		rsp = ts.protocol.Invoke(ctx, pkg)
	} else {
		invokeCtx, cancelFunc := context.WithTimeout(ctx, cfg.HandleTimeout)
		go func() {
			rsp = ts.protocol.Invoke(invokeCtx, pkg)
			cancelFunc()
		}()
		<-invokeCtx.Done()
		if len(rsp) == 0 { // The rsp must be none-empty
			rsp = ts.protocol.InvokeTimeout(pkg)
		}
	}
	return rsp
}
