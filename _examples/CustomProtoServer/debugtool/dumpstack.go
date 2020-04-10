package main

import (
	"fmt"

	"github.com/TarsCloud/TarsGo/tars"
	"github.com/TarsCloud/TarsGo/tars/protocol/res/adminf"
)

func main() {
	comm := tars.NewCommunicator()
	obj := fmt.Sprintf("Test.CustomProtoServer.CustomProtoObj@tcp -h 127.0.0.1 -p 10015 -t 60000")
	app := new(adminf.AdminF)
	comm.StringToProxy(obj, app)
	ret, err := app.Notify("taf.dumpstack")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(ret)
}
