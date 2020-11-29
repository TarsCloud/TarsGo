package main

import (
	"context"
	"fmt"

	"github.com/TarsCloud/TarsGo/tars"

	"StressTest"
)

func main() {
	comm := tars.NewCommunicator()
	obj := fmt.Sprintf("StressTest.ContextTestServer.ContextTestObj@tcp -h 127.0.0.1 -p 10028 -t 60000")
	app := new(StressTest.ContextTest)
	comm.StringToProxy(obj, app)
	var out, i int32
	i = 1111
	ctx := context.Background()
	c := make(map[string]string)
	c["a"] = "b"
	ret, err := app.AddWithContext(ctx, i, i*2, &out, c)
	if code := tars.GetErrorCode(err); code != 0 {
		fmt.Printf("error code: %d, message %v\n", code, err)
		return
	}
	fmt.Println(c)
	fmt.Println(ret, out)
}
