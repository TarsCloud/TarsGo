//服务端启动初始化，解析命令行参数，解析配置

package tars

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/TarsCloud/TarsGo/tars/protocol/res/adminf"
	"github.com/TarsCloud/TarsGo/tars/transport"
	"github.com/TarsCloud/TarsGo/tars/util/conf"
	"github.com/TarsCloud/TarsGo/tars/util/endpoint"
	"github.com/TarsCloud/TarsGo/tars/util/grace"
	"github.com/TarsCloud/TarsGo/tars/util/rogger"
	"github.com/TarsCloud/TarsGo/tars/util/tools"
)

var tarsConfig map[string]*transport.TarsServerConf
var goSvrs map[string]*transport.TarsServer
var httpSvrs map[string]*http.Server
var listenFds []*os.File
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
	//svrCfg.netThread = sMap["netthread"]
	svrCfg.netThread = c.GetInt("/tars/application/server<netthread>")

	svrCfg.log = sMap["log"]
	//add version info
	svrCfg.Version = TarsVersion
	//add adapters config
	svrCfg.Adapters = make(map[string]adapterConfig)

	rogger.SetLevel(rogger.StringToLevel(svrCfg.LogLevel))
	TLOG.SetFileRoller(svrCfg.LogPath+"/"+svrCfg.App+"/"+svrCfg.Server, 10, 100)

	//client
	cltCfg = new(clientConfig)
	cMap := c.GetMap("/tars/application/client")
	cltCfg.Locator = cMap["locator"]
	cltCfg.stat = cMap["stat"]
	cltCfg.property = cMap["property"]
	cltCfg.AsyncInvokeTimeout = c.GetInt("/tars/application/client<async-invoke-timeout>")
	cltCfg.refreshEndpointInterval = c.GetInt("/tars/application/client<refresh-endpoint-interval>")
	serList = c.GetDomain("/tars/application/server")

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
			MaxInvoke:     int32(MaxInvoke),
			AcceptTimeout: AcceptTimeout,
			ReadTimeout:   ReadTimeout,
			WriteTimeout:  WriteTimeout,
			HandleTimeout: HandleTimeout,
			IdleTimeout:   IdleTimeout,

			TCPNoDelay:     TCPNoDelay,
			TCPReadBuffer:  TCPReadBuffer,
			TCPWriteBuffer: TCPWriteBuffer,
		}

		tarsConfig[svrObj] = conf
	}
	TLOG.Debug("config add ", tarsConfig)
	localString := c.GetString("/tars/application/server<local>")
	localpoint := endpoint.Parse(localString)

	adminCfg := &transport.TarsServerConf{
		Proto:          "tcp",
		Address:        fmt.Sprintf("%s:%d", localpoint.Host, localpoint.Port),
		MaxInvoke:      int32(MaxInvoke),
		AcceptTimeout:  AcceptTimeout,
		ReadTimeout:    ReadTimeout,
		WriteTimeout:   WriteTimeout,
		HandleTimeout:  HandleTimeout,
		IdleTimeout:    IdleTimeout,
		QueueCap:       QueueCap,
		TCPNoDelay:     TCPNoDelay,
		TCPReadBuffer:  TCPReadBuffer,
		TCPWriteBuffer: TCPWriteBuffer,
	}

	tarsConfig["AdminObj"] = adminCfg
	svrCfg.Adapters["AdminAdapter"] = adapterConfig{localpoint, "tcp", "AdminObj", 1}
	go initReport()
	go initProReport()
}

//Run the application
func Run() {
	Init()
	<-statInited
	<-proInited

	for _, env := range os.Environ() {
		if strings.HasPrefix(env, grace.InheritFdPrefix) {
			TLOG.Infof("env %s", env)
		}
	}
	// add adminF
	adf := new(adminf.AdminF)
	ad := new(Admin)
	AddServant(adf, ad, "AdminObj")

	for _, obj := range objRunList {
		if s, ok := httpSvrs[obj]; ok {
			go func(obj string) {
				fmt.Println(obj, "http server start")
				addr := s.Addr
				if addr == "" {
					addr = ":http"
				}
				ln, err := grace.CreateListener("tcp", addr)
				if err != nil {
					fmt.Println(obj, "server start failed", err)
					os.Exit(1)
				}
				err = s.Serve(ln)
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
	go reportNotifyInfo("restart")

	if os.Getenv("GRACE_RESTART") == "1" {
		ppid := os.Getppid()
		TLOG.Infof("stop ppid %d", ppid)
		if ppid > 1 {
			grace.SignalUSR2(ppid)
		}
	}
	mainloop()
}

func graceRestart() {
	pid := os.Getpid()
	TLOG.Debugf("grace restart server begin %d", pid)
	os.Setenv("GRACE_RESTART", "1")
	envs := os.Environ()
	newEnvs := make([]string, 0)
	for _, env := range envs {
		// skip fd inherited from parent process
		if strings.HasPrefix(env, grace.InheritFdPrefix) {
			continue
		}
		newEnvs = append(newEnvs, env)
	}

	files := []*os.File{os.Stdin, os.Stdout, os.Stderr}
	for key, file := range grace.GetAllLisenFiles() {
		fd := fmt.Sprint(file.Fd())
		newFd := len(files)
		TLOG.Debugf("tranlate %s=%s to %s=%d", key, fd, key, newFd)
		newEnvs = append(newEnvs, fmt.Sprintf("%s=%d", key, newFd))
		files = append(files, file)
	}

	exePath, err := exec.LookPath(os.Args[0])
	if err != nil {
		TLOG.Errorf("LookPath failed %v", err)
		return
	}

	process, err := os.StartProcess(exePath, os.Args, &os.ProcAttr{
		Env:   newEnvs,
		Files: files,
	})
	if err != nil {
		TLOG.Errorf("start supprocess failed %v", err)
		return
	}
	TLOG.Infof("subprocess start %d", process.Pid)
	go process.Wait()
}

func graceShutdown() {
	pid := os.Getpid()
	TLOG.Infof("grace shutdown start %d", pid)
	timeout := time.Second * 30
	var err error
	ctx, _ := context.WithTimeout(context.Background(), timeout)
	for _, obj := range objRunList {
		if s, ok := httpSvrs[obj]; ok {
			err = s.Shutdown(ctx)
		}
		if s, ok := goSvrs[obj]; ok {
			err = s.Shutdown(ctx)
		}
	}
	if err == nil {
		TLOG.Infof("grace shutdown succ %d", pid)
	} else {
		TLOG.Infof("grace shutdown failed within %d seconds: %v", timeout/time.Second, err)
	}
	shutdown <- true
}

func handleSignal() {
	usrFun, killFunc := graceRestart, graceShutdown
	grace.GraceHandler(usrFun, killFunc)
}

func mainloop() {
	ha := new(NodeFHelper)
	comm := NewCommunicator()
	//comm.SetProperty("netthread", 1)
	node := GetServerConfig().Node
	app := GetServerConfig().App
	server := GetServerConfig().Server
	//container := GetServerConfig().Container
	ha.SetNodeInfo(comm, node, app, server)

	go ha.ReportVersion(GetServerConfig().Version)
	go ha.KeepAlive("") //first start

	go handleSignal()

	loop := time.NewTicker(MainLoopTicker)
	for {
		select {
		case <-shutdown:
			reportNotifyInfo("stop")
			return
		case <-loop.C:
			for name, adapter := range svrCfg.Adapters {
				if adapter.Protocol == "not_tars" {
					//TODO not_tars support
					ha.KeepAlive(name)
					continue
				}
				if s, ok := goSvrs[adapter.Obj]; ok {
					if !s.IsZombie(ZombileTimeout) {
						ha.KeepAlive(name)
					}
				}
			}

		}
	}
}
