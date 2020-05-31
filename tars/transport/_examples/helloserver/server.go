package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/TarsCloud/TarsGo/tars/transport"
)

// MyServer struct for testing tars tcp server
type MyServer struct{}

// Invoke recv request and make response.
func (s *MyServer) Invoke(ctx context.Context, req []byte) (rsp []byte) {
	fmt.Println("recv", string(req))
	rsp = make([]byte, 4)
	rsp = append(rsp, []byte("Hello ")...)
	rsp = append(rsp, req[4:]...)
	binary.BigEndian.PutUint32(rsp[:4], uint32(len(rsp)))
	return
}

// ParsePackage parse package from buff,check if tars package finished.
func (s *MyServer) ParsePackage(buff []byte) (pkgLen, status int) {
	if len(buff) < 4 {
		return 0, transport.PACKAGE_LESS
	}
	length := binary.BigEndian.Uint32(buff[:4])

	if length > 1048576000 || len(buff) > 1048576000 { // 1000MB
		return 0, transport.PACKAGE_ERROR
	}
	if len(buff) < int(length) {
		return 0, transport.PACKAGE_LESS
	}
	return int(length), transport.PACKAGE_FULL
}

// InvokeTimeout how to detail with timeout package.
func (s *MyServer) InvokeTimeout(pkg []byte) []byte {
	payload := []byte("timeout")
	ret := make([]byte, 4+len(payload))
	binary.BigEndian.PutUint32(pkg[:4], uint32(len(ret)))
	copy(pkg[4:], payload)
	return ret
}

func main() {
	conf := &transport.TarsServerConf{
		Proto:         "tcp",
		Address:       "localhost:3333",
		MaxInvoke:     20,
		AcceptTimeout: time.Millisecond * 500,
		ReadTimeout:   time.Millisecond * 100,
		WriteTimeout:  time.Millisecond * 100,
		HandleTimeout: time.Millisecond * 6000,
		IdleTimeout:   time.Millisecond * 600000,
	}
	s := MyServer{}
	svr := transport.NewTarsServer(&s, conf)
	err := svr.Serve()
	if err != nil {
		panic(err)
	}
}
