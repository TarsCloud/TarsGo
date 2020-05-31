package main

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/TarsCloud/TarsGo/tars/transport"
)

// CustomProtocolImp str protocol object implements ServerProtocol interface
type CustomProtocolImp struct {
}

// ParsePackage parse request package
func (s *CustomProtocolImp) ParsePackage(buff []byte) (int, int) {
	if len(buff) < 4 {
		return 0, transport.PACKAGE_LESS
	}
	if len(buff) > 10485760 {
		return 0, transport.PACKAGE_ERROR
	}
	var idx = bytes.Index(buff, []byte("\n"))
	if idx > 0 {
		return idx + 1, transport.PACKAGE_FULL
	}

	return 0, transport.PACKAGE_LESS
}

// Invoke process request and send response 
func (s *CustomProtocolImp) Invoke(ctx context.Context, req []byte) []byte {
	fmt.Println("req:", req)
	reqMap, err := url.ParseQuery(strings.TrimSpace(string(req)))
	if err != nil {
		return []byte("ret=-1&msg=invalid_format\n")
	}

	cmd := reqMap.Get("cmd")
	data := reqMap.Get("data")
	if cmd == "hello" {
		return []byte(fmt.Sprintf("ret=%d&data=hello,%s\n", 0, data))
	} else if cmd == "echo" {
		return []byte(fmt.Sprintf("ret=%d&data=%s\n", 0, data))
	} else {
		return []byte(fmt.Sprintf("ret=%d&data=%s\n", -1, "invalid cmd"))
	}
}

// InvokeTimeout send response when server is timeout
func (s *CustomProtocolImp) InvokeTimeout(pkg []byte) []byte {
	fmt.Println("invoke timeout:", pkg)
	rsp := bytes.NewBuffer(nil)
	rsp.WriteString("ret=-1&data=timeout\n")
	return rsp.Bytes()
}
