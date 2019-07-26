package main

import (
	"tars"

	"_SERVER_/_APP_"
)

func main() {
	cfg := tars.GetServerConfig() // Get Config File Object

	// Init servant
	imp := new(_SERVANT_Imp)
	app := new(_APP_._SERVANT_)
	// Register Servant
	app.AddServant(imp, cfg.App+"."+cfg.Server+"._SERVANT_Obj")

	tars.Run()
}
