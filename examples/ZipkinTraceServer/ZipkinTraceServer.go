package main

import (
	"github.com/TarsCloud/TarsGo/tars"

	"ZipkinTraceServer/tars-protocol/ZipkinTraceApp"
	"github.com/TarsCloud/TarsGo/tars/plugin/zipkintracing"
)

func main() { // Init servant
	imp := new(ZipkinTraceImp)             // New Imp
	app := new(ZipkinTraceApp.ZipkinTrace) // New init the A Tars
	cf := zipkintracing.ZipkinClientFilter()
	sf := zipkintracing.ZipkinServerFilter()
	tars.RegisterClientFilter(cf)
	tars.RegisterServerFilter(sf)

	cfg := tars.GetServerConfig() // Get Config File Object
	zipkintracing.InitV2()
	app.AddServantWithContext(imp, cfg.App+"."+cfg.Server+".ZipkinTraceObj") // Register Servant
	tars.Run()
}
