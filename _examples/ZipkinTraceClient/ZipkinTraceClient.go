package main

import (
	"fmt"
	"strconv"

	"github.com/TarsCloud/TarsGo/tars"
	"github.com/TarsCloud/TarsGo/tars/plugin/zipkintracing"

	"ZipkinTraceApp"
	"net/http"
	_ "net/http/pprof"
)

var sapp = new(ZipkinTraceApp.ZipkinTrace)
var comm = tars.NewCommunicator()
var logger = tars.GetLogger("zipkin")

func main() { //Init servant
	go func() {
		fmt.Println(http.ListenAndServe("0.0.0.0:8090", nil))
	}()
	imp := new(ZipkinClientImp)             //New Imp
	app := new(ZipkinTraceApp.ZipkinClient) //New init the A Tars
	comm.StringToProxy("ZipkinTraceApp.ZipkinTraceServer.ZipkinTraceObj", sapp)
	cf := zipkintracing.ZipkinClientFilter()
	sf := zipkintracing.ZipkinServerFilter()
	tars.RegisterClientFilter(cf)
	tars.RegisterServerFilter(sf)
	cfg := tars.GetServerConfig() //Get Config File Object
	port := cfg.Adapters[cfg.App+"."+cfg.Server+".ZipkinClientObjAdapter"].Endpoint.Port
	hostPort := cfg.Adapters[cfg.App+"."+cfg.Server+".ZipkinClientObjAdapter"].Endpoint.Host + ":" + strconv.Itoa(int(port))
	zipkintracing.Init("http://127.0.0.1:9411/api/v2/spans", true, true, true, hostPort, cfg.App+"."+cfg.Server)
	app.AddServantWithContext(imp, cfg.App+"."+cfg.Server+".ZipkinClientObj") //Register Servant
	tars.Run()
}
