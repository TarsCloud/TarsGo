package main

import (
	"github.com/TarsCloud/TarsGo/tars"

	"ZipkinTraceServer/tars-protocol/ZipkinTraceApp"

	"github.com/TarsCloud/TarsGo/contrib/middleware/zipkintracing"
)

func main() { // Init servant
	cfg := tars.GetServerConfig() // Get Config File Object

	zipkintracing.Init()
	tars.UseClientFilterMiddleware(zipkintracing.ZipkinClientFilter())
	tars.UseServerFilterMiddleware(zipkintracing.ZipkinServerFilter())

	imp := new(ZipkinTraceImp)                                               // New Imp
	app := new(ZipkinTraceApp.ZipkinTrace)                                   // New init the A Tars
	app.AddServantWithContext(imp, cfg.App+"."+cfg.Server+".ZipkinTraceObj") // Register Servant
	tars.Run()
}
