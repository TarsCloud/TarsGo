package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"

	"github.com/TarsCloud/TarsGo/contrib/middleware/zipkintracing"
	"github.com/TarsCloud/TarsGo/tars"

	"ZipkinTraceClient/tars-protocol/ZipkinTraceApp"
)

var (
	sapp   = new(ZipkinTraceApp.ZipkinTrace)
	logger = tars.GetLogger("zipkin")
)

func main() { // Init servant
	go func() {
		fmt.Println(http.ListenAndServe("0.0.0.0:8090", nil))
	}()
	cfg := tars.GetServerConfig() // Get Config File Object

	zipkintracing.Init()
	tars.UseClientFilterMiddleware(zipkintracing.ZipkinClientFilter())
	tars.UseServerFilterMiddleware(zipkintracing.ZipkinServerFilter())

	comm := tars.GetCommunicator()
	comm.StringToProxy("ZipkinTraceApp.ZipkinTraceServer.ZipkinTraceObj@tcp -h 127.0.0.1 -p 15015 -t 60000", sapp)
	imp := new(ZipkinClientImp)                                               // New Imp
	app := new(ZipkinTraceApp.ZipkinClient)                                   // New init the A Tars
	app.AddServantWithContext(imp, cfg.App+"."+cfg.Server+".ZipkinClientObj") // Register Servant
	tars.Run()
}
