package transport

import (
	"context"
	"net"
)

const (
	// PackageLess shows is not a completed package.
	PackageLess = iota
	// PackageFull shows is a completed package.
	PackageFull
	// PackageError shows is a error package.
	PackageError
)

// ServerHandler is interface with listen and handler method
type ServerHandler interface {
	Listen() error
	Handle() error
	OnShutdown()
	CloseIdles(n int64) bool
}

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
	OnClose(address string)
	OnConnect(address string)
}

func isNoDataError(err error) bool {
	netErr, ok := err.(net.Error)
	if ok && netErr.Timeout() && netErr.Temporary() {
		return true
	}
	return false
}
