package main

import (
	"fmt"
	"github.com/TarsCloud/TarsGo/tars"
	"github.com/TarsCloud/TarsGo/tars/plugin/zipkintracing"

	"ZipkinTraceClient/tars-protocol/ZipkinTraceApp"
	"net/http"
	_ "net/http/pprof"
)

var sapp = new(ZipkinTraceApp.ZipkinTrace)
var comm = tars.NewCommunicator()
var logger = tars.GetLogger("zipkin")

func main() { // Init servant
	go func() {
		fmt.Println(http.ListenAndServe("0.0.0.0:8090", nil))
	}()
	imp := new(ZipkinClientImp)             // New Imp
	app := new(ZipkinTraceApp.ZipkinClient) // New init the A Tars
	comm.StringToProxy("ZipkinTraceApp.ZipkinTraceServer.ZipkinTraceObj@tcp -h 127.0.0.1 -p 15015 -t 60000", sapp)
	cf := zipkintracing.ZipkinClientFilter()
	sf := zipkintracing.ZipkinServerFilter()
	tars.RegisterClientFilter(cf)
	tars.RegisterServerFilter(sf)
	cfg := tars.GetServerConfig() // Get Config File Object
	zipkintracing.InitV2()
	app.AddServantWithContext(imp, cfg.App+"."+cfg.Server+".ZipkinClientObj") // Register Servant
	tars.Run()
}
