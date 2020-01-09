package main

import (
	"log"
	"os"

	"github.com/TarsCloud/TarsGo/tars"

	"StressTest"
	"runtime/pprof"
)

func main() { //Init servant
	f, err := os.Create("/usr/local/app/tars/app_log/cpu.profile")
	if err != nil {
		log.Fatal("could not create CPU profile: ", err)
	}
	if err := pprof.StartCPUProfile(f); err != nil {
		log.Fatal("could not start CPU profile: ", err)
	}
	defer pprof.StopCPUProfile()

	imp := new(EchoTestImp)                                    //New Imp
	app := new(StressTest.EchoTest)                            //New init the A JCE
	cfg := tars.GetServerConfig()                              //Get Config File Object
	app.AddServant(imp, cfg.App+"."+cfg.Server+".EchoTestObj") //Register Servant
	tars.Run()
}
