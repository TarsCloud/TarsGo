package main

import (
	"github.com/TarsCloud/TarsGo/tars"

	"StressTest"
	"log"
	"net/http"
	_ "net/http/pprof"
)

func main() { //Init servant
	go func() {
		log.Fatal(http.ListenAndServe("0.0.0.0:8090", nil))
	}()

	imp := new(ContextTestImp)                                               //New Imp
	app := new(StressTest.ContextTest)                                       //New init the A Tars
	cfg := tars.GetServerConfig()                                            //Get Config File Object
	app.AddServantWithContext(imp, cfg.App+"."+cfg.Server+".ContextTestObj") //Register Servant
	tars.Run()
}
