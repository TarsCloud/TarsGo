//服务端启动初始化，解析命令行参数，解析配置

package tars

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/TarsCloud/TarsGo/tars/protocol"
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
var isShudowning int32

// TLOG is the logger for tars framework.
var TLOG = rogger.GetLogger("TLOG")
var initOnce sync.Once
var shutdownOnce sync.Once

type adminFn func(string) (string, error)

var adminMethods map[string]adminFn
var destroyableObjs []destroyableImp

type destroyableImp interface {
	Destroy()
}

func init() {
	tarsConfig = make(map[string]*transport.TarsServerConf)
	goSvrs = make(map[string]*transport.TarsServer)
	httpSvrs = make(map[string]*http.Server)
	shutdown = make(chan bool, 1)
	adminMethods = make(map[string]adminFn)
	rogger.SetLevel(rogger.ERROR)
}

// ServerConfigPath is the path of server config
var ServerConfigPath string

func initConfig() {
	if ServerConfigPath == "" {
		svrFlag := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
		svrFlag.StringVar(&ServerConfigPath, "config", "", "server config path")
		svrFlag.Parse(os.Args[1:])
	}

	if len(ServerConfigPath) == 0 {
		return
	}

	c, err := conf.NewConf(ServerConfigPath)
	if err != nil {
		TLOG.Errorf("Parse server config fail %v", err)
		return
	}

	//Config.go
	//Server
	svrCfg = new(serverConfig)
	if strings.EqualFold(c.GetString("/tars/application<enableset>"), "Y") {
		svrCfg.Enableset = true
		svrCfg.Setdivision = c.GetString("/tars/application<setdivision>")
	}
	sMap := c.GetMap("/tars/application/server")
	svrCfg.Node = sMap["node"]
	svrCfg.App = sMap["app"]
	svrCfg.Server = sMap["server"]
	svrCfg.LocalIP = sMap["localip"]
	svrCfg.Local = c.GetString("/tars/application/server<local>")
	//svrCfg.Container = c.GetString("/tars/application<container>")

	//init log
	svrCfg.LogPath = sMap["logpath"]
	svrCfg.LogSize = tools.ParseLogSizeMb(sMap["logsize"])
	svrCfg.LogNum = tools.ParseLogNum(sMap["lognum"])
	svrCfg.LogLevel = sMap["logLevel"]
	svrCfg.Config = sMap["config"]
	svrCfg.Notify = sMap["notify"]
	svrCfg.BasePath = sMap["basepath"]
	svrCfg.DataPath = sMap["datapath"]
	svrCfg.Log = sMap["log"]

	//add version info
	svrCfg.Version = TarsVersion
	//add adapters config
	svrCfg.Adapters = make(map[string]adapterConfig)

	cachePath := filepath.Join(svrCfg.DataPath, svrCfg.Server) + ".tarsdat"
	if cacheData, err := ioutil.ReadFile(cachePath); err == nil {
		json.Unmarshal(cacheData, &appCache)
	}

	if svrCfg.LogLevel == "" {
		svrCfg.LogLevel = appCache.LogLevel
	} else {
		appCache.LogLevel = svrCfg.LogLevel
	}
	rogger.SetLevel(rogger.StringToLevel(svrCfg.LogLevel))
	if svrCfg.LogPath != "" {
		TLOG.SetFileRoller(svrCfg.LogPath+"/"+svrCfg.App+"/"+svrCfg.Server, 10, 100)
	}

	//cache
	appCache.TarsVersion = TarsVersion

	// add timeout config
	svrCfg.AcceptTimeout = tools.ParseTimeOut(c.GetIntWithDef("/tars/application/server<accepttimeout>", AcceptTimeout))
	svrCfg.ReadTimeout = tools.ParseTimeOut(c.GetIntWithDef("/tars/application/server<readtimeout>", ReadTimeout))
	svrCfg.WriteTimeout = tools.ParseTimeOut(c.GetIntWithDef("/tars/application/server<writetimeout>", WriteTimeout))
	svrCfg.HandleTimeout = tools.ParseTimeOut(c.GetIntWithDef("/tars/application/server<handletimeout>", HandleTimeout))
	svrCfg.IdleTimeout = tools.ParseTimeOut(c.GetIntWithDef("/tars/application/server<idletimeout>", IdleTimeout))
	svrCfg.ZombileTimeout = tools.ParseTimeOut(c.GetIntWithDef("/tars/application/server<zombiletimeout>", ZombileTimeout))
	svrCfg.QueueCap = c.GetIntWithDef("/tars/application/server<queuecap>", QueueCap)
	svrCfg.GracedownTimeout = tools.ParseTimeOut(c.GetIntWithDef("/tars/application/server<gracedowntimeout>", GracedownTimeout))

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
	svrCfg.StatReportChannelBufLen = c.GetInt32WithDef("/tars/application/server<statreportchannelbuflen>", StatReportChannelBufLen)
	// maxPackageLength
	svrCfg.MaxPackageLength = c.GetIntWithDef("/tars/application/server<maxPackageLength>", MaxPackageLength)
	protocol.SetMaxPackageLength(svrCfg.MaxPackageLength)

	//client
	cltCfg = new(clientConfig)
	cMap := c.GetMap("/tars/application/client")
	cltCfg.Locator = cMap["locator"]
	cltCfg.Stat = cMap["stat"]
	cltCfg.Property = cMap["property"]
	cltCfg.AsyncInvokeTimeout = c.GetIntWithDef("/tars/application/client<async-invoke-timeout>", AsyncInvokeTimeout)
	cltCfg.RefreshEndpointInterval = c.GetIntWithDef("/tars/application/client<refresh-endpoint-interval>", refreshEndpointInterval)
	serList = c.GetDomain("/tars/application/server")
	cltCfg.ReportInterval = c.GetIntWithDef("/tars/application/client<report-interval>", reportInterval)
	cltCfg.CheckStatusInterval = c.GetIntWithDef("/tars/application/client<check-status-interval>", checkStatusInterval)

	// add client timeout
	cltCfg.ClientQueueLen = c.GetIntWithDef("/tars/application/client<clientqueuelen>", ClientQueueLen)
	cltCfg.ClientIdleTimeout = tools.ParseTimeOut(c.GetIntWithDef("/tars/application/client<clientidletimeout>", ClientIdleTimeout))
	cltCfg.ClientReadTimeout = tools.ParseTimeOut(c.GetIntWithDef("/tars/application/client<clientreadtimeout>", ClientReadTimeout))
	cltCfg.ClientWriteTimeout = tools.ParseTimeOut(c.GetIntWithDef("/tars/application/client<clientwritetimeout>", ClientWriteTimeout))
	cltCfg.ClientDialTimeout = tools.ParseTimeOut(c.GetIntWithDef("/tars/application/client<clientdialtimeout>", ClientDialTimeout))
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

	if len(svrCfg.Local) > 0 {
		localpoint := endpoint.Parse(svrCfg.Local)
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
		RegisterAdmin(rogger.Admin, rogger.HandleDyeingAdmin)
	}

	go initReport()
}

// Run the application
func Run() {
	defer rogger.FlushLogger()
	isShudowning = 0
	Init()
	<-statInited

	for _, env := range os.Environ() {
		if strings.HasPrefix(env, grace.InheritFdPrefix) {
			TLOG.Infof("env %s", env)
		}
	}

	// add adminF
	if _, ok := tarsConfig["AdminObj"]; ok {
		adf := new(adminf.AdminF)
		ad := new(Admin)
		AddServant(adf, ad, "AdminObj")
	}

	lisDone := &sync.WaitGroup{}
	for _, obj := range objRunList {
		if s, ok := httpSvrs[obj]; ok {
			lisDone.Add(1)
			go func(obj string) {
				addr := s.Addr
				TLOG.Infof("%s http server start on %s", obj, s.Addr)
				if addr == "" {
					lisDone.Done()
					teerDown(fmt.Errorf("empty addr for %s", obj))
					return
				}
				ln, err := grace.CreateListener("tcp", addr)
				if err != nil {
					lisDone.Done()
					teerDown(fmt.Errorf("start http server for %s failed: %v", obj, err))
					return
				}

				lisDone.Done()
				err = s.Serve(ln)
				if err != nil {
					if err == http.ErrServerClosed {
						TLOG.Infof("%s http server stop: %v", obj, err)
					} else {
						teerDown(fmt.Errorf("%s server stop: %v", obj, err))
					}
				}
			}(obj)
			continue
		}

		s := goSvrs[obj]
		if s == nil {
			teerDown(fmt.Errorf("Obj not found %s", obj))
			break
		}
		TLOG.Debugf("Run %s  %+v", obj, s.GetConfig())
		lisDone.Add(1)
		go func(obj string) {
			if err := s.Listen(); err != nil {
				lisDone.Done()
				teerDown(fmt.Errorf("listen obj for %s failed: %v", obj, err))
				return
			}

			lisDone.Done()
			if err := s.Serve(); err != nil {
				teerDown(fmt.Errorf("server obj for %s failed: %v", obj, err))
				return
			}
		}(obj)
	}
	go ReportNotifyInfo(NOTIFY_NORMAL, "restart")

	lisDone.Wait()
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

	// redirect stdout/stderr to logger
	cfg := GetServerConfig()
	var logfile *os.File
	if cfg != nil {
		GetLogger("")
		logpath := filepath.Join(cfg.LogPath, cfg.App, cfg.Server, cfg.App+"."+cfg.Server+".log")
		logfile, _ = os.OpenFile(logpath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
		TLOG.Debugf("redirect to %s %v", logpath, logfile)
	}
	if logfile == nil {
		logfile = os.Stdout
	}
	files := []*os.File{os.Stdin, logfile, logfile}
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
	var wg sync.WaitGroup

	atomic.StoreInt32(&isShudowning, 1)
	pid := os.Getpid()

	var graceShutdownTimeout time.Duration
	if atomic.LoadInt32(&isShutdownbyadmin) == 1 {
		// shutdown by admin,we should need shorten the timeout
		graceShutdownTimeout = tools.ParseTimeOut(GracedownTimeout)
	} else {
		graceShutdownTimeout = svrCfg.GracedownTimeout
	}

	TLOG.Infof("grace shutdown start %d in %v", pid, graceShutdownTimeout)
	ctx, cancel := context.WithTimeout(context.Background(), graceShutdownTimeout)

	for _, obj := range destroyableObjs {
		wg.Add(1)
		go func(wg *sync.WaitGroup, obj destroyableImp) {
			defer wg.Done()
			obj.Destroy()
			TLOG.Infof("grace Destroy succ %d", pid)
		}(&wg, obj)
	}

	for _, obj := range objRunList {
		if s, ok := httpSvrs[obj]; ok {
			wg.Add(1)
			go func(s *http.Server, ctx context.Context, wg *sync.WaitGroup, objstr string) {
				defer wg.Done()
				err := s.Shutdown(ctx)
				if err == nil {
					TLOG.Infof("grace shutdown http %s succ %d", objstr, pid)
				} else {
					TLOG.Infof("grace shutdown http %s failed within %v : %v", objstr, graceShutdownTimeout, err)
				}
			}(s, ctx, &wg, obj)
		}

		if s, ok := goSvrs[obj]; ok {
			wg.Add(1)
			go func(s *transport.TarsServer, ctx context.Context, wg *sync.WaitGroup, objstr string) {
				defer wg.Done()
				err := s.Shutdown(ctx)
				if err == nil {
					TLOG.Infof("grace shutdown tars %s succ %d", objstr, pid)
				} else {
					TLOG.Infof("grace shutdown tars %s failed within %v: %v", objstr, graceShutdownTimeout, err)
				}
			}(s, ctx, &wg, obj)
		}
	}

	go func() {
		wg.Wait()
		cancel()
	}()

	select {
	case <-ctx.Done():
		TLOG.Infof("grace shutdown all succ within : %v", graceShutdownTimeout)
	case <-time.After(graceShutdownTimeout):
		TLOG.Infof("grace shutdown timeout within : %v", graceShutdownTimeout)
	}

	teerDown(nil)
}

func teerDown(err error) {
	shutdownOnce.Do(func() {
		if err != nil {
			ReportNotifyInfo(NOTIFY_NORMAL, "server is fatal: "+err.Error())
			fmt.Println(err)
			TLOG.Error(err)
		}
		shutdown <- true
	})
}

func handleSignal() {
	usrFun, killFunc := graceRestart, graceShutdown
	grace.GraceHandler(usrFun, killFunc)
}

func mainloop() {
	ha := new(NodeFHelper)
	comm := NewCommunicator()
	node := GetServerConfig().Node
	app := GetServerConfig().App
	server := GetServerConfig().Server
	container := GetServerConfig().Container
	ha.SetNodeInfo(comm, node, app, server, container)

	go ha.ReportVersion(GetServerConfig().Version)
	go ha.KeepAlive("") //first start
	go handleSignal()
	loop := time.NewTicker(GetServerConfig().MainLoopTicker)

	for {
		select {
		case <-shutdown:
			ReportNotifyInfo(NOTIFY_NORMAL, "stop")
			return
		case <-loop.C:
			if atomic.LoadInt32(&isShudowning) == 1 {
				continue
			}
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
