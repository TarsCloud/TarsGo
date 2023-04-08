package tars

import (
	"crypto/tls"

	"github.com/TarsCloud/TarsGo/tars/transport"
)

type ServerConfOption func(*transport.TarsServerConf)

func WithQueueCap(queueCap int) ServerConfOption {
	return func(c *transport.TarsServerConf) {
		c.QueueCap = queueCap
	}
}

func WithTlsConfig(tlsConfig *tls.Config) ServerConfOption {
	return func(c *transport.TarsServerConf) {
		c.TlsConfig = tlsConfig
	}
}

func WithMaxInvoke(maxInvoke int32) ServerConfOption {
	return func(c *transport.TarsServerConf) {
		c.MaxInvoke = maxInvoke
	}
}

func newTarsServerConf(proto, address string, svrCfg *serverConfig, opts ...ServerConfOption) *transport.TarsServerConf {
	tarsSvrConf := &transport.TarsServerConf{
		Proto:          proto,
		Address:        address,
		MaxInvoke:      svrCfg.MaxInvoke,
		AcceptTimeout:  svrCfg.AcceptTimeout,
		ReadTimeout:    svrCfg.ReadTimeout,
		WriteTimeout:   svrCfg.WriteTimeout,
		HandleTimeout:  svrCfg.HandleTimeout,
		IdleTimeout:    svrCfg.IdleTimeout,
		QueueCap:       svrCfg.QueueCap,
		TCPNoDelay:     svrCfg.TCPNoDelay,
		TCPReadBuffer:  svrCfg.TCPReadBuffer,
		TCPWriteBuffer: svrCfg.TCPWriteBuffer,
	}
	for _, opt := range opts {
		opt(tarsSvrConf)
	}
	return tarsSvrConf
}
