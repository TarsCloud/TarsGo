package main

import (
	"context"
	"fmt"
	"github.com/TarsCloud/TarsGo/tars"
	"github.com/TarsCloud/TarsGo/tars/util/current"

	"trace/frontend/tars-protocol/Trace"
)

func main() {
	tars.ServerConfigPath = "config.conf"
	comm := tars.GetCommunicator()
	obj := fmt.Sprintf("Trace.TarsTraceFrontServer.FrontendObj@tcp -h 127.0.0.1 -p 10015 -t 60000")
	app := new(Trace.Frontend)
	comm.StringToProxy(obj, app)
	var out, i int32
	i = 123
	ctx := current.ContextWithTarsCurrent(context.Background())
	current.OpenTarsTrace(ctx, true)
	ret, err := app.AddWithContext(ctx, i, i*2, &out)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(ret, out)
}
