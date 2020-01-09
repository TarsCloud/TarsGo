package main

import (
	"StressTest"

	_ "net/http/pprof"

	"github.com/TarsCloud/TarsGo/tars"
	"github.com/TarsCloud/TarsGo/tars/util/rogger"
)

var app *StressTest.EchoTest
var preport *tars.PropertyReport
var log *rogger.Logger
var remoteLog *rogger.Logger
var info string
var comm *tars.Communicator

//EchoClientImp struct
type EchoClientImp struct {
}

//Add implement
func (imp *EchoClientImp) Add(a int32, b int32, c *int32) (int32, error) {
	sum := tars.NewSum()
	count := tars.NewCount()
	preport = tars.CreatePropertyReport("testproperty", sum, count)
	*c = a + b
	strIn := make([]int8, 100, 100)
	for i := 0; i < 100; i++ {
		strIn[i] = int8(i)
	}

	var strOut []int8
	app := new(StressTest.EchoTest)
	obj := "StressTest.EchoTestServer.EchoTestObj"
	comm.StringToProxy(obj, app)
	_, err := app.Echo(strIn, &strOut)
	if err != nil {
		log.Error("call error: ", err)

	} else {
		log.Debug("call succss and result[10] is ", strOut[10])

	}
	remoteLog.Info(info)
	preport.Report(int(*c))
	return 0, nil
}

//Sub implement
func (imp *EchoClientImp) Sub(a int32, b int32, c *int32) (int32, error) {
	*c = a - b
	return 0, nil
}

func main() { //Init servant
	comm = tars.NewCommunicator()

	// client
	log = tars.GetDayLogger("report", 1)
	//config
	cfg := tars.GetServerConfig()
	remoteConf := tars.NewRConf(cfg.App, cfg.Server, cfg.BasePath)
	config, _ := remoteConf.GetConfig("test.conf")
	info = config

	remoteLog = tars.GetRemoteLogger("configstring")

	// server
	imp := new(EchoClientImp)          //New Imp
	apps := new(StressTest.EchoClient) //New init the A Tars
	//cfg := tars.GetServerConfig()      //Get Config File Object
	apps.AddServant(imp, cfg.App+"."+cfg.Server+".EchoClientObj")
	tars.Run()
}
