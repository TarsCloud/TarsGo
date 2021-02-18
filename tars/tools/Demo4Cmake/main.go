package main

import (
	"fmt"
	"os"

	"github.com/TarsCloud/TarsGo/tars"

	"_MODULE_/tars-protocol/_APP_"
)

func main() {
	// Get server config
	cfg := tars.GetServerConfig()

	// New servant imp
	imp := new(_SERVANT_Imp)
	err := imp.Init()
	if err != nil {
		fmt.Printf("_SERVANT_Imp init fail, err:(%s)\n", err)
		os.Exit(-1)
	}
	// New servant
	app := new(_APP_._SERVANT_)
	// Register Servant
	app.AddServantWithContext(imp, cfg.App+"."+cfg.Server+"._SERVANT_Obj")
	
	// Run application
	tars.Run()
}
