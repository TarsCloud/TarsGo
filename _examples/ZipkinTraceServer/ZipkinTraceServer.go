package main

import (
	"github.com/TarsCloud/TarsGo/tars"

	"ZipkinTraceApp"
	"strconv"

	"github.com/TarsCloud/TarsGo/tars/plugin/zipkintracing"
)

func main() { //Init servant
	imp := new(ZipkinTraceImp)             //New Imp
	app := new(ZipkinTraceApp.ZipkinTrace) //New init the A Tars
	cf := zipkintracing.ZipkinClientFilter()
	sf := zipkintracing.ZipkinServerFilter()
	tars.RegisterClientFilter(cf)
	tars.RegisterServerFilter(sf)

	cfg := tars.GetServerConfig() //Get Config File Object
	port := cfg.Adapters[cfg.App+"."+cfg.Server+".ZipkinTraceObjAdapter"].Endpoint.Port
	hostPort := cfg.Adapters[cfg.App+"."+cfg.Server+".ZipkinTraceObjAdapter"].Endpoint.Host + ":" + strconv.Itoa(int(port))
	zipkintracing.Init("http://127.0.0.1:9411/api/v2/spans", true, true, true, hostPort, cfg.App+"."+cfg.Server)
	app.AddServantWithContext(imp, cfg.App+"."+cfg.Server+".ZipkinTraceObj") //Register Servant
	tars.Run()
}
