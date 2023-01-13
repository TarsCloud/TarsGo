package main

import (
	"fmt"
	"time"

	"github.com/TarsCloud/TarsGo/tars"
	"github.com/TarsCloud/TarsGo/tars/protocol/push"
)

func callback(data []byte) {
	fmt.Println("recv message:", string(data))
}

func main() {
	comm := tars.GetCommunicator()
	obj := "TestApp.PushServer.MessageObj@tcp -h 127.0.0.1 -p 10015 -t 60000"
	client := push.NewClient(callback)
	comm.StringToProxy(obj, client)
	data, err := client.Connect([]byte("hello"))
	if err != nil {
		panic(err)
	}
	fmt.Println("connect ok", string(data))
	// Wait for receving message
	time.Sleep(time.Second * 10)
}
