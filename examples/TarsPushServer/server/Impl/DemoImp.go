package Impl

import (
	"TarPushServer/demo"
	"context"
	"fmt"
	"github.com/TarsCloud/TarsGo/tars/util/current"
	"strconv"
)

var app = &demo.DemoObj{}

type DemoImp struct {
}

func GetApp() *demo.DemoObj {
	return app
}
func (d DemoImp) Reg(ctx context.Context, req *demo.RegReq, rsp *demo.RegRsp) (err error) {
	rsp.Msg = req.Msg
	go func() {
		for i := 0; i < 10; i++ {
			msg := fmt.Sprintf("push msg %d", i)
			context := make(map[string]string, 1)
			context["msg"] = "******" + strconv.Itoa(i)
			uuid, _ := current.GetUUID(ctx)
			context["uuid"] = uuid
			GetApp().AsyncSendResponse_Push(ctx, &msg, context)
		}
	}()
	return nil
}

func (d DemoImp) Push(ctx context.Context, msg *string) (err error) {
	return nil
}

func (d DemoImp) Invoke(ctx context.Context, pkg []byte) []byte {
	//TODO implement me
	fmt.Println("implement me")
	return []byte{}
}

func (d DemoImp) GetCloseMsg() []byte {
	//TODO implement me
	fmt.Println("implement me")
	return nil
}

func (d DemoImp) DoClose(ctx context.Context) {
	//TODO implement me
	fmt.Println("implement me")
}
