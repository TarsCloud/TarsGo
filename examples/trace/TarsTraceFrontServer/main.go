package main

import (
	"fmt"
	"os"

	"github.com/TarsCloud/TarsGo/tars"

	"trace/frontend/tars-protocol/Trace"
)

func main() {
	// Get server config
	cfg := tars.GetServerConfig()

	// New servant imp
	imp := new(FrontendImp)
	err := imp.Init()
	if err != nil {
		fmt.Printf("FrontendImp init fail, err:(%s)\n", err)
		os.Exit(-1)
	}
	// New servant
	app := new(Trace.Frontend)
	// Register Servant
	app.AddServantWithContext(imp, cfg.App+"."+cfg.Server+".FrontendObj")

	// Run application
	tars.Run()
}
