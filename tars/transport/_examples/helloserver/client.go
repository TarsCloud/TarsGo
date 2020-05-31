package main

import (
	"encoding/binary"
	"fmt"
	"strconv"
	"time"

	"github.com/TarsCloud/TarsGo/tars/transport"
)

// MyClient is a example client for tars client testing.
type MyClient struct {
	recvCount int
}

// Recv print pkg and count
func (c *MyClient) Recv(pkg []byte) {
	fmt.Println("recv", string(pkg))
	c.recvCount++
}

// ParsePackage parse package from buff
func (c *MyClient) ParsePackage(buff []byte) (pkgLen, status int) {
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

func getMsg(name string) []byte {
	payload := []byte(name)
	pkg := make([]byte, 4+len(payload))
	binary.BigEndian.PutUint32(pkg[:4], uint32(len(pkg)))
	copy(pkg[4:], payload)
	return pkg
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
	client := transport.NewTarsClient("localhost:3333", cp, conf)

	name := "Bob"
	count := 500
	for i := 0; i < count; i++ {
		msg := getMsg(name + strconv.Itoa(i))
		err := client.Send(msg)
		if err != nil {
			fmt.Println("send err: " + err.Error())
		}
	}

	time.Sleep(time.Second * 1)
	if count != cp.recvCount {
		fmt.Println("bad")
	} else {
		fmt.Println("good")
	}
	client.Close()
}
