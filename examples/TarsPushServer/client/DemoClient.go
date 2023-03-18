package main

import (
	"TarPushServer/demo"
	"fmt"
	"github.com/TarsCloud/TarsGo/tars"
	"time"
)

var prx *demo.DemoObj

type Callback struct {
	start int64
	cost  int64
	count int64
}

func (c *Callback) Reg_Callback(ret *demo.Result, rsp *demo.RegRsp, opt ...map[string]string) {
	//TODO implement me
	panic("implement me")
}

func (c *Callback) Notify_Callback(ret *demo.Result, opt ...map[string]string) {
	//TODO implement me
	panic("implement me")
}

func (c *Callback) Notify_ExceptionCallback(err error) {
	//TODO implement me
	panic("implement me")
}

func (c *Callback) Push_Callback(msg *string, opt ...map[string]string) {
	/*	if c.count == 0 {
			c.start = time.Now().UnixMicro()
		}
		c.count++
		if c.count == 500000 {
			c.cost = time.Now().UnixMicro() - c.start
			fmt.Printf("cost--->%vus\n", c.cost)
		}*/

	//if c.start == 0 {
	//	c.start = time.Now().UnixMicro()
	//} else {
	//	c.cost = time.Now().UnixMicro() - c.start
	//}
	//fmt.Printf("cost--->%vus|Push---->:%s======%v\n", c.cost, *msg, opt)
	fmt.Printf("%v|Push---->:%s======%v\n", time.Now().UnixMilli(), *msg, opt)
}

func (c Callback) Push_ExceptionCallback(err error) {
	panic("implement me")
}

func (c Callback) Reg_ExceptionCallback(err error) {
	//TODO implement me
	panic("implement me")
}

func TestReg() {
	req := &demo.RegReq{Msg: "reg"}
	rsp := &demo.RegRsp{}
	prx.Reg(req, rsp)
}

func main() {
	com := tars.GetCommunicator()
	obj := "Base.DemoServer.DemoObj@tcp -h 127.0.0.1 -p 8888 -t 60000"
	prx = &demo.DemoObj{}
	com.StringToProxy(obj, prx)
	prx.SetOnConnectCallback(func(s string) {
		fmt.Println("<-----------onConnect--------->")
		TestReg()
	})
	prx.SetOnCloseCallback(func(s string) {
		fmt.Println("<-----------onClose----------->")
	})
	TarsCb := new(Callback)
	prx.SetTarsCallback(TarsCb)
	go func() {
		ticker := time.NewTicker(time.Second * 10)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				prx.TarsPing()
			}
		}
	}()
	prx.TarsPing()
	for true {
		time.Sleep(time.Second * 10)
	}
}
