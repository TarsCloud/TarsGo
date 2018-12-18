//服务端启动初始化，解析命令行参数，解析配置

package tars

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/TarsCloud/TarsGo/tars/protocol/res/adminf"
	"github.com/TarsCloud/TarsGo/tars/transport"
	"github.com/TarsCloud/TarsGo/tars/util/conf"
	"github.com/TarsCloud/TarsGo/tars/util/endpoint"
	"github.com/TarsCloud/TarsGo/tars/util/rogger"
	"github.com/TarsCloud/TarsGo/tars/util/tools"
)

var tarsConfig map[string]*transport.TarsServerConf
var goSvrs map[string]*transport.TarsServer
var httpSvrs map[string]*http.Server
var shutdown chan bool
var serList []string
var objRunList []string

//TLOG is the logger for tars framework.
var TLOG = rogger.GetLogger("TLOG")
var initOnce sync.Once

type adminFn func(string) (string, error)

var adminMethods map[string]adminFn

func init() {
	tarsConfig = make(map[string]*transport.TarsServerConf)
	goSvrs = make(map[string]*transport.TarsServer)
	httpSvrs = make(map[string]*http.Server)
	shutdown = make(chan bool, 1)
	adminMethods = make(map[string]adminFn)
	rogger.SetLevel(rogger.ERROR)
	Init()
}

//Init should run before GetServerConfig & GetClientConfig , or before run
// and Init should be only run once
func Init() {
	initOnce.Do(initConfig)
}

func initConfig() {
	confPath := flag.String("config", "", "init config path")
	flag.Parse()
	if len(*confPath) == 0 {
		return
	}
	c, err := conf.NewConf(*confPath)
	if err != nil {
		TLOG.Error("open app config fail")
	}
	//Config.go
	//Server
	svrCfg = new(serverConfig)
	if c.GetString("/tars/application<enableset>") == "Y" {
		svrCfg.Enableset = true
		svrCfg.Setdivision = c.GetString("/tars/application<setdivision>")
	}
	sMap := c.GetMap("/tars/application/server")
	svrCfg.Node = sMap["node"]
	svrCfg.App = sMap["app"]
	svrCfg.Server = sMap["server"]
	svrCfg.LocalIP = sMap["localip"]
	//svrCfg.Container = c.GetString("/tars/application<container>")
	//init log
	svrCfg.LogPath = sMap["logpath"]
	svrCfg.LogSize = tools.ParseLogSizeMb(sMap["logsize"])
	svrCfg.LogNum = tools.ParseLogNum(sMap["lognum"])
	svrCfg.LogLevel = sMap["logLevel"]
	svrCfg.config = sMap["config"]
	svrCfg.notify = sMap["notify"]
	svrCfg.BasePath = sMap["basepath"]
	svrCfg.DataPath = sMap["datapath"]

	svrCfg.log = sMap["log"]
	//add version info
	svrCfg.Version = TarsVersion
	//add adapters config
	svrCfg.Adapters = make(map[string]adapterConfig)

	rogger.SetLevel(rogger.StringToLevel(svrCfg.LogLevel))
	TLOG.SetFileRoller(svrCfg.LogPath+"/"+svrCfg.App+"/"+svrCfg.Server, 10, 100)

	// add timeout config
	svrCfg.AcceptTimeout = tools.ParseTimeOut(c.GetIntWithDef("/tars/application/server<accepttimeout>", AcceptTimeout))
	svrCfg.ReadTimeout = tools.ParseTimeOut(c.GetIntWithDef("/tars/application/server<readtimeout>", ReadTimeout))
	svrCfg.WriteTimeout = tools.ParseTimeOut(c.GetIntWithDef("/tars/application/server<writetimeout>", WriteTimeout))
	svrCfg.HandleTimeout = tools.ParseTimeOut(c.GetIntWithDef("/tars/application/server<handletimeout>", HandleTimeout))
	svrCfg.IdleTimeout = tools.ParseTimeOut(c.GetIntWithDef("/tars/application/server<idletimeout>", IdleTimeout))
	svrCfg.ZombileTimeout = tools.ParseTimeOut(c.GetIntWithDef("/tars/application/server<zombiletimeout>", ZombileTimeout))
	svrCfg.QueueCap = c.GetIntWithDef("/tars/application/server<queuecap>", QueueCap)

	// add tcp config
	svrCfg.TCPReadBuffer = c.GetIntWithDef("/tars/application/server<tcpreadbuffer>", TCPReadBuffer)
	svrCfg.TCPWriteBuffer = c.GetIntWithDef("/tars/application/server<tcpwritebuffer>", TCPWriteBuffer)
	svrCfg.TCPNoDelay = c.GetBoolWithDef("/tars/application/server<tcpnodelay>", TCPNoDelay)
	// add routine number
	svrCfg.MaxInvoke = c.GetInt32WithDef("/tars/application/server<maxroutine>", MaxInvoke)
	// add adapter & report config
	svrCfg.PropertyReportInterval = tools.ParseTimeOut(c.GetIntWithDef("/tars/application/server<propertyreportinterval>", PropertyReportInterval))
	svrCfg.StatReportInterval = tools.ParseTimeOut(c.GetIntWithDef("/tars/application/server<statreportinterval>", StatReportInterval))
	svrCfg.MainLoopTicker = tools.ParseTimeOut(c.GetIntWithDef("/tars/application/server<mainloopticker>", MainLoopTicker))

	//client
	cltCfg = new(clientConfig)
	cMap := c.GetMap("/tars/application/client")
	cltCfg.Locator = cMap["locator"]
	cltCfg.stat = cMap["stat"]
	cltCfg.property = cMap["property"]
	cltCfg.AsyncInvokeTimeout = c.GetIntWithDef("/tars/application/client<async-invoke-timeout>", AsyncInvokeTimeout)
	cltCfg.refreshEndpointInterval = c.GetIntWithDef("/tars/application/client<refresh-endpoint-interval>", refreshEndpointInterval)
	serList = c.GetDomain("/tars/application/server")
	cltCfg.reportInterval = c.GetIntWithDef("/tars/application/client<report-interval>", reportInterval)

	// add client timeout
	cltCfg.ClientQueueLen = c.GetIntWithDef("/tars/application/client<clientqueuelen>", ClientQueueLen)
	cltCfg.ClientIdleTimeout = tools.ParseTimeOut(c.GetIntWithDef("/tars/application/client<clientidletimeout>", ClientIdleTimeout))
	cltCfg.ClientReadTimeout = tools.ParseTimeOut(c.GetIntWithDef("/tars/application/client<clientreadtimeout>", ClientReadTimeout))
	cltCfg.ClientWriteTimeout = tools.ParseTimeOut(c.GetIntWithDef("/tars/application/client<clientwritetimeout>", ClientWriteTimeout))
	cltCfg.ReqDefaultTimeout = c.GetInt32WithDef("/tars/application/client<reqdefaulttimeout>", ReqDefaultTimeout)
	cltCfg.ObjQueueMax = c.GetInt32WithDef("/tars/application/client<objqueuemax>", ObjQueueMax)
	cltCfg.AdapterProxyTicker = tools.ParseTimeOut(c.GetIntWithDef("/tars/application/client<adapterproxyticker>", AdapterProxyTicker))
	cltCfg.AdapterProxyResetCount = c.GetIntWithDef("/tars/application/client<adapterproxyresetcount>", AdapterProxyResetCount)


	for _, adapter := range serList {
		endString := c.GetString("/tars/application/server/" + adapter + "<endpoint>")
		end := endpoint.Parse(endString)
		svrObj := c.GetString("/tars/application/server/" + adapter + "<servant>")
		protocol := c.GetString("/tars/application/server/" + adapter + "<protocol>")
		threads := c.GetInt("/tars/application/server/" + adapter + "<threads>")
		svrCfg.Adapters[adapter] = adapterConfig{end, protocol, svrObj, threads}
		host := end.Host
		if end.Bind != "" {
			host = end.Bind
		}
		conf := &transport.TarsServerConf{
			Proto:         end.Proto,
			Address:       fmt.Sprintf("%s:%d", host, end.Port),
			MaxInvoke:     svrCfg.MaxInvoke,
			AcceptTimeout: svrCfg.AcceptTimeout,
			ReadTimeout:   svrCfg.ReadTimeout,
			WriteTimeout:  svrCfg.WriteTimeout,
			HandleTimeout: svrCfg.HandleTimeout,
			IdleTimeout:   svrCfg.IdleTimeout,

			TCPNoDelay:     svrCfg.TCPNoDelay,
			TCPReadBuffer:  svrCfg.TCPReadBuffer,
			TCPWriteBuffer: svrCfg.TCPWriteBuffer,
		}

		tarsConfig[svrObj] = conf
	}
	TLOG.Debug("config add ", tarsConfig)
	localString := c.GetString("/tars/application/server<local>")
	localpoint := endpoint.Parse(localString)

	adminCfg := &transport.TarsServerConf{
		Proto:          "tcp",
		Address:        fmt.Sprintf("%s:%d", localpoint.Host, localpoint.Port),
		MaxInvoke:      svrCfg.MaxInvoke,
		AcceptTimeout:  svrCfg.AcceptTimeout,
		ReadTimeout:    svrCfg.ReadTimeout,
		WriteTimeout:   svrCfg.WriteTimeout,
		HandleTimeout:  svrCfg.HandleTimeout,
		IdleTimeout:    svrCfg.IdleTimeout,
		QueueCap:       svrCfg.QueueCap,
		TCPNoDelay:     svrCfg.TCPNoDelay,
		TCPReadBuffer:  svrCfg.TCPReadBuffer,
		TCPWriteBuffer: svrCfg.TCPWriteBuffer,
	}

	tarsConfig["AdminObj"] = adminCfg
	svrCfg.Adapters["AdminAdapter"] = adapterConfig{localpoint, "tcp", "AdminObj", 1}
	go initReport()
}

//Run the application
func Run() {
	Init()
	<-statInited

	// add adminF
	adf := new(adminf.AdminF)
	ad := new(Admin)
	AddServant(adf, ad, "AdminObj")

	for _, obj := range objRunList {
		if s, ok := httpSvrs[obj]; ok {
			go func(obj string) {
				fmt.Println(obj, "http server start")
				err := s.ListenAndServe()
				if err != nil {
					fmt.Println(obj, "server start failed", err)
					os.Exit(1)
				}
			}(obj)
			continue
		}

		s := goSvrs[obj]
		if s == nil {
			TLOG.Debug("Obj not found", obj)
			break
		}
		TLOG.Debug("Run", obj, s.GetConfig())
		go func(obj string) {
			err := s.Serve()
			if err != nil {
				fmt.Println(obj, "server start failed", err)
				os.Exit(1)
			}
		}(obj)
	}
	go ReportNotifyInfo("restart")
	mainloop()
}

func mainloop() {
	ha := new(NodeFHelper)
	comm := NewCommunicator()
	node := GetServerConfig().Node
	app := GetServerConfig().App
	server := GetServerConfig().Server
	//container := GetServerConfig().Container
	ha.SetNodeInfo(comm, node, app, server)

	go ha.ReportVersion(GetServerConfig().Version)
	go ha.KeepAlive("") //first start
	loop := time.NewTicker(GetServerConfig().MainLoopTicker)
	for {
		select {
		case <-shutdown:
			ReportNotifyInfo("stop")
			return
		case <-loop.C:
			for name, adapter := range svrCfg.Adapters {
				if adapter.Protocol == "not_tars" {
					//TODO not_tars support
					ha.KeepAlive(name)
					continue
				}
				if s, ok := goSvrs[adapter.Obj]; ok {
					if !s.IsZombie(GetServerConfig().ZombileTimeout) {
						ha.KeepAlive(name)
					}
				}
			}

		}
	}
}
