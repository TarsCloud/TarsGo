package transport

import (
	"context"
	"net"
)

const (
	// PACKAGE_LESS shows is not a completed package.
	PACKAGE_LESS = iota
	// PACKAGE_FULL shows is a completed package.
	PACKAGE_FULL
	// PACKAGE_ERROR shows is a error package.
	PACKAGE_ERROR
)

// ServerProtocol is interface for handling the server side tars package.
type ServerProtocol interface {
	Invoke(ctx context.Context, pkg []byte) []byte
	ParsePackage(buff []byte) (int, int)
	InvokeTimeout(pkg []byte) []byte
	GetCloseMsg() []byte
	DoClose(ctx context.Context)
}

// ClientProtocol interface for handling tars client package.
type ClientProtocol interface {
	Recv(pkg []byte)
	ParsePackage(buff []byte) (int, int)
}

func isNoDataError(err error) bool {
	netErr, ok := err.(net.Error)
	if ok && netErr.Timeout() && netErr.Temporary() {
		return true
	}
	return false
}
