package main

import (
	"fmt"

	"github.com/TarsCloud/TarsGo/tars"

	"StressTest"
)

func main() {
	comm := tars.NewCommunicator()
	obj := fmt.Sprintf("StressTest.EchoClientServer.EchoClientObj@tcp -h 127.0.0.1 -p 10015  -t 60000")
	app := new(StressTest.EchoClient)
	comm.StringToProxy(obj, app)
	var out, i int32
	for i = 0; i < 100; i++ {
		ret, err := app.Add(i, i*2, &out)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(ret, out)
	}
}
