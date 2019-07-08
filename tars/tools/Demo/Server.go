package main

import (
	"tars"

	"_APP_"
)

func main() { //Init servant
	imp := new(_SERVANT_Imp)                                    //New Imp
	app := new(_APP_._SERVANT_)                                 //New init the A Tars
	cfg := tars.GetServerConfig()                               //Get Config File Object
	app.AddServant(imp, cfg.App+"."+cfg.Server+"._SERVANT_Obj") //Register Servant
	tars.Run()
}
