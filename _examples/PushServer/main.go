package main

import (
	"context"
	"fmt"
	"time"

	"github.com/TarsCloud/TarsGo/tars"
	"github.com/TarsCloud/TarsGo/tars/protocol/push"
	"github.com/TarsCloud/TarsGo/tars/util/current"
)

type pushImp struct{}

// OnConnect ...
func (p *pushImp) OnConnect(ctx context.Context, req []byte) []byte {
	ip, _ := current.GetClientIPFromContext(ctx)
	port, _ := current.GetClientPortFromContext(ctx)
	fmt.Println("on connect:", ip, port)
	go func() {
		for i := 0; i < 3; i++ {
			time.Sleep(time.Millisecond * 100)
			if err := push.Send(ctx, []byte("msg"+fmt.Sprint(i))); err != nil {
				fmt.Println("send error", err)
			}
		}
	}()
	return req
}

// OnClose ...
func (p *pushImp) OnClose(ctx context.Context) {
	ip, _ := current.GetClientIPFromContext(ctx)
	port, _ := current.GetClientPortFromContext(ctx)
	fmt.Println("on close:", ip, port)
}

func main() {
	cfg := tars.GetServerConfig()
	proto := push.NewServer(&pushImp{})
	tars.AddServantWithProtocol(proto, cfg.App+"."+cfg.Server+".MessageObj")
	tars.Run()
}
