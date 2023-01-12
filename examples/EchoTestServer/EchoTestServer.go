package main

import (
	"log"
	"os"
	"path"

	"github.com/TarsCloud/TarsGo/tars"

	"EchoTestServer/tars-protocol/StressTest"
	"runtime/pprof"
)

func main() { //Init servant
	cfg := tars.GetServerConfig() //Get Config File Object
	f, err := os.Create(path.Join(cfg.LogPath, "cpu.profile"))
	if err != nil {
		log.Fatal("could not create CPU profile: ", err)
	}
	if err := pprof.StartCPUProfile(f); err != nil {
		log.Fatal("could not start CPU profile: ", err)
	}
	defer pprof.StopCPUProfile()

	imp := new(EchoTestImp)                                    //New Imp
	app := new(StressTest.EchoTest)                            //New init the A JCE
	app.AddServant(imp, cfg.App+"."+cfg.Server+".EchoTestObj") //Register Servant
	tars.Run()
}
