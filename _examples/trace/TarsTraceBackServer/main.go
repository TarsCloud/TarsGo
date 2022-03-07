package main

import (
	"fmt"
	"os"

	"github.com/TarsCloud/TarsGo/tars"

	"trace/backend/tars-protocol/Trace"
)

func main() {
	// Get server config
	cfg := tars.GetServerConfig()

	// New servant imp
	imp := new(BackendImp)
	err := imp.Init()
	if err != nil {
		fmt.Printf("BackendImp init fail, err:(%s)\n", err)
		os.Exit(-1)
	}
	// New servant
	app := new(Trace.Backend)
	// Register Servant
	app.AddServantWithContext(imp, cfg.App+"."+cfg.Server+".BackendObj")

	// Run application
	tars.Run()
}
