package main

import (
	"fmt"

	"github.com/TarsCloud/TarsGo/tars"

	"TlsTestServer/tars-protocol/App"
)

func main() {
	tars.ServerConfigPath = "config.conf"
	comm := tars.NewCommunicator()
	obj := fmt.Sprintf("App.TlsTestServer.TlsObj@ssl -h 127.0.0.1 -p 13015 -t 6000000000")
	app := new(App.Tls)
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
