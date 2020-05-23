package main

import (
	"fmt"

	"github.com/TarsCloud/TarsGo/tars"

	"CmakeServer/tars-protocol/StressTest"
)

func main() {
	comm := tars.NewCommunicator()
	obj := fmt.Sprintf("StressTest.EchoTestServer.EchoTestObj@tcp -h 127.0.0.1 -p 10015 -t 60000")
	app := new(StressTest.EchoTest)
	comm.StringToProxy(obj, app)
	var out, i []int8
	i = []int8{5, 2, 0}
	ret, err := app.Echo(i, &out)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(ret, out)
}
