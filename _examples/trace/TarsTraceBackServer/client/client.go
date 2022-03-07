package main

import (
	"fmt"

	"github.com/TarsCloud/TarsGo/tars"

	"trace/backend/tars-protocol/Trace"
)

func main() {
	comm := tars.NewCommunicator()
	obj := fmt.Sprintf("Trace.TarsTraceBackServer.BackendObj@tcp -h 127.0.0.1 -p 20015 -t 60000")
	app := new(Trace.Backend)
	comm.StringToProxy(obj, app)
	var out, i int32
	i = 123
	ret, err := app.Add(i, i*2, &out)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(ret, out)
}
