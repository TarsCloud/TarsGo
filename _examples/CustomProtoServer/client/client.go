package main

import (
	"bytes"
	"fmt"
	"strconv"
	"time"

	"github.com/TarsCloud/TarsGo/tars/transport"
)

type MyClient struct {
	recvCount int
}

func (c *MyClient) Recv(pkg []byte) {
	fmt.Println("recv:", string(pkg))
	c.recvCount++
}
func (c *MyClient) ParsePackage(buff []byte) (pkgLen, status int) {
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

func getMsg(name string, idx int) []byte {
	arr := []string{"hello", "echo"}
	msg := []byte(fmt.Sprintf("cmd=%s&data=%s\n", arr[idx%2], name+strconv.Itoa(idx)))
	return msg
}

func main() {
	cp := &MyClient{}
	conf := &transport.TarsClientConf{
		Proto:        "tcp",
		QueueLen:     10000,
		IdleTimeout:  time.Second * 5,
		ReadTimeout:  time.Millisecond * 100,
		WriteTimeout: time.Millisecond * 1000,
	}
	client := transport.NewTarsClient("localhost:10015", cp, conf)

	name := "Bob"
	count := 100
	for i := 0; i < count; i++ {
		msg := getMsg(name, i)
		client.Send(msg)
	}

	time.Sleep(time.Second * 2)
	if count != cp.recvCount {
		fmt.Println("bad")
	} else {
		fmt.Println("good")
	}
	client.Close()
}
