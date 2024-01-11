// 服务端启动初始化，解析命令行参数，解析配置

package tars

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/automaxprocs/maxprocs"

	"github.com/TarsCloud/TarsGo/tars/protocol"
	"github.com/TarsCloud/TarsGo/tars/protocol/res/adminf"
	"github.com/TarsCloud/TarsGo/tars/transport"
	"github.com/TarsCloud/TarsGo/tars/util/conf"
	"github.com/TarsCloud/TarsGo/tars/util/endpoint"
	"github.com/TarsCloud/TarsGo/tars/util/grace"
	"github.com/TarsCloud/TarsGo/tars/util/rogger"
	"github.com/TarsCloud/TarsGo/tars/util/ssl"
	"github.com/TarsCloud/TarsGo/tars/util/tools"
)

type destroyableImp interface {
	Destroy()
}

type application struct {
	conf               *conf.Conf
	opt                *options
	svrCfg             *serverConfig
	cltCfg             *clientConfig
	communicator       *Communicator
	onceCommunicator   sync.Once
	tarsConfig         map[string]*transport.TarsServerConf
	goSvrs             map[string]*transport.TarsServer
	httpSvrs           map[string]*http.Server
	serList            []string
	objRunList         []string
	clientObjInfo      map[string]map[string]string
	clientObjTlsConfig map[string]*tls.Config
	clientTlsConfig    *tls.Config

	rConf     *RConf
	onceRConf sync.Once

	appCache         AppCache
	destroyableObjs  []destroyableImp
	adminMethods     map[string]adminFn
	allFilters       *filters
	dispatchReporter DispatchReporter

	shutdown          chan bool
	isShutdownByAdmin int32
	isShutdowning     int32
	shutdownOnce      sync.Once
	initOnce          sync.Once
}

var (
	// TLOG is the logger for tars framework.
	TLOG = rogger.GetLogger("TLOG")

	defaultApp       *application
	ServerConfigPath string
)

func init() {
	rogger.SetLevel(rogger.ERROR)

	defaultApp = newApp()
}

func newApp() *application {
	return &application{
		opt:                &options{},
		cltCfg:             newClientConfig(),
		svrCfg:             newServerConfig(),
		tarsConfig:         make(map[string]*transport.TarsServerConf),
		goSvrs:             make(map[string]*transport.TarsServer),
		httpSvrs:           make(map[string]*http.Server),
		clientObjInfo:      make(map[string]map[string]string),
		clientObjTlsConfig: make(map[string]*tls.Config),
		adminMethods:       make(map[string]adminFn),
		shutdown:           make(chan bool, 1),
		allFilters:         &filters{},
	}
}

// GetConf Get server conf.Conf config
func GetConf() *conf.Conf {
	return defaultApp.GetConf()
}

func Run(opts ...Option) {
	defaultApp.Run(opts...)
}

// GetConf Get server conf.Conf config
func (a *application) GetConf() *conf.Conf {
	a.init()
	return a.conf
}

func (a *application) init() {
	a.initOnce.Do(func() {
		a.initConfig()
	})
}

func (a *application) initConfig() {
	defer func() {
		go func() {
			_ = statInitOnce.Do(func() error {
				return initReport(a)
			})
		}()
	}()
	if ServerConfigPath == "" {
		svrFlag := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
		svrFlag.StringVar(&ServerConfigPath, "config", "", "server config path")
		_ = svrFlag.Parse(os.Args[1:])
	}

	if len(ServerConfigPath) == 0 {
		return
	}

	c, err := conf.NewConf(ServerConfigPath)
	if err != nil {
		TLOG.Errorf("Parse server config fail %v", err)
		return
	}

	// parse config
	a.parseServerConfig(c)
	a.parseClientConfig(c)
	a.conf = c

	cachePath := filepath.Join(a.svrCfg.DataPath, a.svrCfg.Server) + ".tarsdat"
	if cacheData, err := os.ReadFile(cachePath); err == nil {
		_ = json.Unmarshal(cacheData, &a.appCache)
	}
	// cache
	a.appCache.TarsVersion = Version
	if a.svrCfg.LogLevel == "" {
		a.svrCfg.LogLevel = a.appCache.LogLevel
	} else {
		a.appCache.LogLevel = a.svrCfg.LogLevel
	}
	rogger.SetLevel(rogger.StringToLevel(a.svrCfg.LogLevel))
	if a.svrCfg.LogPath != "" {
		_ = TLOG.SetFileRoller(a.svrCfg.LogPath+"/"+a.svrCfg.App+"/"+a.svrCfg.Server, int(a.svrCfg.LogNum), int(a.svrCfg.LogSize))
	}
	protocol.SetMaxPackageLength(a.svrCfg.MaxPackageLength)
	_, _ = maxprocs.Set(maxprocs.Logger(TLOG.Infof))
}

func (a *application) parseServerConfig(c *conf.Conf) {
	// init server config
	if strings.EqualFold(c.GetString("/tars/application<enableset>"), "Y") {
		a.svrCfg.Enableset = true
		a.svrCfg.Setdivision = c.GetString("/tars/application<setdivision>")
	}
	sMap := c.GetMap("/tars/application/server")
	a.svrCfg.Node = sMap["node"]
	a.svrCfg.App = sMap["app"]
	a.svrCfg.Server = sMap["server"]
	a.svrCfg.LocalIP = c.GetStringWithDef("/tars/application/server<localip>", a.svrCfg.LocalIP)
	a.svrCfg.NodeName = c.GetStringWithDef("/tars/application/server<node_name>", a.svrCfg.LocalIP)
	a.svrCfg.Local = c.GetString("/tars/application/server<local>")
	// svrCfg.Container = c.GetString("/tars/application<container>")

	// init log
	a.svrCfg.LogPath = sMap["logpath"]
	a.svrCfg.LogSize = tools.ParseLogSizeMb(sMap["logsize"])
	a.svrCfg.LogNum = tools.ParseLogNum(sMap["lognum"])
	a.svrCfg.LogLevel = sMap["logLevel"]
	a.svrCfg.Config = sMap["config"]
	a.svrCfg.Notify = sMap["notify"]
	a.svrCfg.BasePath = sMap["basepath"]
	a.svrCfg.DataPath = sMap["datapath"]
	a.svrCfg.Log = sMap["log"]

	// add version info
	a.svrCfg.Version = Version
	// add adapters config
	a.svrCfg.Adapters = make(map[string]adapterConfig)

	// add timeout config
	a.svrCfg.AcceptTimeout = tools.ParseTimeOut(c.GetIntWithDef("/tars/application/server<accepttimeout>", AcceptTimeout))
	a.svrCfg.ReadTimeout = tools.ParseTimeOut(c.GetIntWithDef("/tars/application/server<readtimeout>", ReadTimeout))
	a.svrCfg.WriteTimeout = tools.ParseTimeOut(c.GetIntWithDef("/tars/application/server<writetimeout>", WriteTimeout))
	a.svrCfg.HandleTimeout = tools.ParseTimeOut(c.GetIntWithDef("/tars/application/server<handletimeout>", HandleTimeout))
	a.svrCfg.IdleTimeout = tools.ParseTimeOut(c.GetIntWithDef("/tars/application/server<idletimeout>", IdleTimeout))
	a.svrCfg.ZombieTimeout = tools.ParseTimeOut(c.GetIntWithDef("/tars/application/server<zombietimeout>", ZombieTimeout))
	a.svrCfg.QueueCap = c.GetIntWithDef("/tars/application/server<queuecap>", QueueCap)
	a.svrCfg.GracedownTimeout = tools.ParseTimeOut(c.GetIntWithDef("/tars/application/server<gracedowntimeout>", GracedownTimeout))

	// add tcp config
	a.svrCfg.TCPReadBuffer = c.GetIntWithDef("/tars/application/server<tcpreadbuffer>", TCPReadBuffer)
	a.svrCfg.TCPWriteBuffer = c.GetIntWithDef("/tars/application/server<tcpwritebuffer>", TCPWriteBuffer)
	a.svrCfg.TCPNoDelay = c.GetBoolWithDef("/tars/application/server<tcpnodelay>", TCPNoDelay)
	// add routine number
	a.svrCfg.MaxInvoke = c.GetInt32WithDef("/tars/application/server<maxroutine>", MaxInvoke)
	// add adapter & report config
	a.svrCfg.PropertyReportInterval = tools.ParseTimeOut(c.GetIntWithDef("/tars/application/server<propertyreportinterval>", PropertyReportInterval))
	a.svrCfg.StatReportInterval = tools.ParseTimeOut(c.GetIntWithDef("/tars/application/server<statreportinterval>", StatReportInterval))
	a.svrCfg.MainLoopTicker = tools.ParseTimeOut(c.GetIntWithDef("/tars/application/server<mainloopticker>", MainLoopTicker))
	a.svrCfg.StatReportChannelBufLen = c.GetInt32WithDef("/tars/application/server<statreportchannelbuflen>", StatReportChannelBufLen)
	// maxPackageLength
	a.svrCfg.MaxPackageLength = c.GetIntWithDef("/tars/application/server<maxPackageLength>", MaxPackageLength)

	// tls
	a.svrCfg.Key = c.GetString("/tars/application/server<key>")
	a.svrCfg.Cert = c.GetString("/tars/application/server<cert>")
	var (
		tlsConfig *tls.Config
		err       error
	)
	if a.svrCfg.Key != "" && a.svrCfg.Cert != "" {
		a.svrCfg.CA = c.GetString("/tars/application/server<ca>")
		a.svrCfg.VerifyClient = c.GetStringWithDef("/tars/application/server<verifyclient>", "0") != "0"
		a.svrCfg.Ciphers = c.GetString("/tars/application/server<ciphers>")
		tlsConfig, err = ssl.NewServerTlsConfig(a.svrCfg.CA, a.svrCfg.Cert, a.svrCfg.Key, a.svrCfg.VerifyClient, a.svrCfg.Ciphers)
		if err != nil {
			panic(err)
		}
	}
	a.svrCfg.SampleRate = c.GetFloatWithDef("/tars/application/server<samplerate>", 0)
	a.svrCfg.SampleType = c.GetString("/tars/application/server<sampletype>")
	a.svrCfg.SampleAddress = c.GetString("/tars/application/server<sampleaddress>")
	a.svrCfg.SampleEncoding = c.GetStringWithDef("/tars/application/server<sampleencoding>", "json")

	serList := c.GetDomain("/tars/application/server")
	for _, adapter := range serList {
		endString := c.GetString("/tars/application/server/" + adapter + "<endpoint>")
		end := endpoint.Parse(endString)
		svrObj := c.GetString("/tars/application/server/" + adapter + "<servant>")
		proto := c.GetString("/tars/application/server/" + adapter + "<protocol>")
		queuecap := c.GetIntWithDef("/tars/application/server/"+adapter+"<queuecap>", a.svrCfg.QueueCap)
		threads := c.GetInt("/tars/application/server/" + adapter + "<threads>")
		a.svrCfg.Adapters[adapter] = adapterConfig{end, proto, svrObj, threads}
		host := end.Host
		if end.Bind != "" {
			host = end.Bind
		}
		var opts []ServerConfOption
		opts = append(opts, WithQueueCap(queuecap))
		if end.IsSSL() {
			key := c.GetString("/tars/application/server/" + adapter + "<key>")
			cert := c.GetString("/tars/application/server/" + adapter + "<cert>")
			if key != "" && cert != "" {
				ca := c.GetString("/tars/application/server/" + adapter + "<ca>")
				verifyClient := c.GetString("/tars/application/server/"+adapter+"<verifyclient>") != "0"
				ciphers := c.GetString("/tars/application/server/" + adapter + "<ciphers>")
				var adpTlsConfig *tls.Config
				adpTlsConfig, err = ssl.NewServerTlsConfig(ca, cert, key, verifyClient, ciphers)
				if err != nil {
					panic(err)
				}
				opts = append(opts, WithTlsConfig(adpTlsConfig))
			} else {
				// common tls.Config
				opts = append(opts, WithTlsConfig(tlsConfig))
			}
		}
		a.tarsConfig[svrObj] = newTarsServerConf(end.Proto, fmt.Sprintf("%s:%d", host, end.Port), a.svrCfg, opts...)
	}
	a.serList = serList

	if len(a.svrCfg.Local) > 0 {
		localPoint := endpoint.Parse(a.svrCfg.Local)
		// 管理端口不启动协程池
		a.tarsConfig["AdminObj"] = newTarsServerConf(localPoint.Proto, fmt.Sprintf("%s:%d", localPoint.Host, localPoint.Port), a.svrCfg, WithMaxInvoke(0))
		a.svrCfg.Adapters["AdminAdapter"] = adapterConfig{localPoint, localPoint.Proto, "AdminObj", 1}
		RegisterAdmin(rogger.Admin, rogger.HandleDyeingAdmin)
	}

	TLOG.Debug("config add ", a.tarsConfig)
}

func (a *application) parseClientConfig(c *conf.Conf) {
	// init client config
	cMap := c.GetMap("/tars/application/client")
	a.cltCfg.Locator = cMap["locator"]
	a.cltCfg.Stat = cMap["stat"]
	a.cltCfg.Property = cMap["property"]
	a.cltCfg.ModuleName = cMap["modulename"]
	a.cltCfg.AsyncInvokeTimeout = c.GetIntWithDef("/tars/application/client<async-invoke-timeout>", AsyncInvokeTimeout)
	a.cltCfg.RefreshEndpointInterval = c.GetIntWithDef("/tars/application/client<refresh-endpoint-interval>", refreshEndpointInterval)
	a.cltCfg.ReportInterval = c.GetIntWithDef("/tars/application/client<report-interval>", reportInterval)
	a.cltCfg.CheckStatusInterval = c.GetIntWithDef("/tars/application/client<check-status-interval>", checkStatusInterval)
	a.cltCfg.KeepAliveInterval = c.GetIntWithDef("/tars/application/client<keep-alive-interval>", keepAliveInterval)

	// add client timeout
	a.cltCfg.ClientQueueLen = c.GetIntWithDef("/tars/application/client<clientqueuelen>", ClientQueueLen)
	a.cltCfg.ClientIdleTimeout = tools.ParseTimeOut(c.GetIntWithDef("/tars/application/client<clientidletimeout>", ClientIdleTimeout))
	a.cltCfg.ClientReadTimeout = tools.ParseTimeOut(c.GetIntWithDef("/tars/application/client<clientreadtimeout>", ClientReadTimeout))
	a.cltCfg.ClientWriteTimeout = tools.ParseTimeOut(c.GetIntWithDef("/tars/application/client<clientwritetimeout>", ClientWriteTimeout))
	a.cltCfg.ClientDialTimeout = tools.ParseTimeOut(c.GetIntWithDef("/tars/application/client<clientdialtimeout>", ClientDialTimeout))
	a.cltCfg.ReqDefaultTimeout = c.GetInt32WithDef("/tars/application/client<reqdefaulttimeout>", ReqDefaultTimeout)
	a.cltCfg.ObjQueueMax = c.GetInt32WithDef("/tars/application/client<objqueuemax>", ObjQueueMax)
	a.cltCfg.context["node_name"] = a.svrCfg.NodeName
	ca := c.GetString("/tars/application/client<ca>")
	if ca != "" {
		cert := c.GetString("/tars/application/client<cert>")
		key := c.GetString("/tars/application/client<key>")
		ciphers := c.GetString("/tars/application/client<ciphers>")
		clientTlsConfig, err := ssl.NewClientTlsConfig(ca, cert, key, ciphers)
		if err != nil {
			panic(err)
		}
		a.clientTlsConfig = clientTlsConfig
	}

	auths := c.GetDomain("/tars/application/client")
	for _, objName := range auths {
		authInfo := make(map[string]string)
		// authInfo["accesskey"] = c.GetString("/tars/application/client/" + objName + "<accesskey>")
		// authInfo["secretkey"] = c.GetString("/tars/application/client/" + objName + "<secretkey>")
		authInfo["ca"] = c.GetString("/tars/application/client/" + objName + "<ca>")
		authInfo["cert"] = c.GetString("/tars/application/client/" + objName + "<cert>")
		authInfo["key"] = c.GetString("/tars/application/client/" + objName + "<key>")
		authInfo["ciphers"] = c.GetString("/tars/application/client/" + objName + "<ciphers>")
		a.clientObjInfo[objName] = authInfo
		if authInfo["ca"] != "" {
			objTlsConfig, err := ssl.NewClientTlsConfig(authInfo["ca"], authInfo["cert"], authInfo["key"], authInfo["ciphers"])
			if err != nil {
				panic(err)
			}
			a.clientObjTlsConfig[objName] = objTlsConfig
		}
	}
}

// Run the application
func (a *application) Run(opts ...Option) {
	defer rogger.FlushLogger()
	a.isShutdowning = 0
	a.init()
	<-statInited

	for _, opt := range opts {
		opt(a.opt)
	}

	for _, env := range os.Environ() {
		if strings.HasPrefix(env, grace.InheritFdPrefix) {
			TLOG.Infof("env %s", env)
		}
	}

	// add adminF
	if _, ok := a.tarsConfig["AdminObj"]; ok {
		adf := new(adminf.AdminF)
		ad := newAdmin(a)
		AddServant(adf, ad, "AdminObj")
	}

	lisDone := &sync.WaitGroup{}
	for _, obj := range a.objRunList {
		if s, ok := a.httpSvrs[obj]; ok {
			lisDone.Add(1)
			go func(obj string) {
				addr := s.Addr
				TLOG.Infof("%s http server start on %s", obj, s.Addr)
				if addr == "" {
					lisDone.Done()
					a.teerDown(fmt.Errorf("empty addr for %s", obj))
					return
				}
				ln, err := grace.CreateListener("tcp", addr)
				if err != nil {
					lisDone.Done()
					a.teerDown(fmt.Errorf("start http server for %s failed: %v", obj, err))
					return
				}

				lisDone.Done()
				if s.TLSConfig != nil {
					err = s.ServeTLS(ln, "", "")
				} else {
					err = s.Serve(ln)
				}
				if err != nil {
					if err == http.ErrServerClosed {
						TLOG.Infof("%s http server stop: %v", obj, err)
					} else {
						a.teerDown(fmt.Errorf("%s server stop: %v", obj, err))
					}
				}
			}(obj)
			continue
		}

		s := a.goSvrs[obj]
		if s == nil {
			a.teerDown(fmt.Errorf("obj not found %s", obj))
			break
		}
		TLOG.Debugf("Run %s  %+v", obj, s.GetConfig())
		lisDone.Add(1)
		go func(obj string) {
			if err := s.Listen(); err != nil {
				lisDone.Done()
				a.teerDown(fmt.Errorf("listen obj for %s failed: %v", obj, err))
				return
			}

			lisDone.Done()
			if err := s.Serve(); err != nil {
				a.teerDown(fmt.Errorf("server obj for %s failed: %v", obj, err))
				return
			}
		}(obj)
	}
	go ReportNotifyInfo(NotifyNormal, "restart")

	lisDone.Wait()
	if os.Getenv("GRACE_RESTART") == "1" {
		ppid := os.Getppid()
		TLOG.Infof("stop ppid %d", ppid)
		if ppid > 1 {
			grace.SignalUSR2(ppid)
		}
	}
	a.mainLoop()
}

func (a *application) graceRestart() {
	pid := os.Getpid()
	TLOG.Debugf("grace restart server begin %d", pid)
	_ = os.Setenv("GRACE_RESTART", "1")
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
	svrCfg := a.ServerConfig()
	var logfile *os.File
	if svrCfg != nil {
		GetLogger("")
		logpath := filepath.Join(svrCfg.LogPath, svrCfg.App, svrCfg.Server, svrCfg.App+"."+svrCfg.Server+".log")
		logfile, _ = os.OpenFile(logpath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
		TLOG.Debugf("redirect to %s %v", logpath, logfile)
	}
	if logfile == nil {
		logfile = os.Stdout
	}
	files := []*os.File{os.Stdin, logfile, logfile}
	for key, file := range grace.GetAllListenFiles() {
		fd := fmt.Sprint(file.Fd())
		newFd := len(files)
		TLOG.Debugf("translate %s=%s to %s=%d", key, fd, key, newFd)
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
		TLOG.Errorf("start subprocess failed %v", err)
		return
	}
	TLOG.Infof("subprocess start %d", process.Pid)
	go process.Wait()
}

func (a *application) graceShutdown() {
	var wg sync.WaitGroup

	atomic.StoreInt32(&a.isShutdowning, 1)
	pid := os.Getpid()

	var graceShutdownTimeout time.Duration
	if atomic.LoadInt32(&a.isShutdownByAdmin) == 1 {
		// shutdown by admin,we should need shorten the timeout
		graceShutdownTimeout = tools.ParseTimeOut(GracedownTimeout)
	} else {
		graceShutdownTimeout = a.svrCfg.GracedownTimeout
	}

	TLOG.Infof("grace shutdown start %d in %v", pid, graceShutdownTimeout)
	ctx, cancel := context.WithTimeout(context.Background(), graceShutdownTimeout)
	// deregister service
	wg.Add(1)
	go func() {
		defer wg.Done()
		a.deregisterAdapters(ctx)
	}()

	for _, obj := range a.destroyableObjs {
		wg.Add(1)
		go func(wg *sync.WaitGroup, obj destroyableImp) {
			defer wg.Done()
			obj.Destroy()
			TLOG.Infof("grace Destroy success %d", pid)
		}(&wg, obj)
	}

	for _, obj := range a.objRunList {
		if s, ok := a.httpSvrs[obj]; ok {
			wg.Add(1)
			go func(s *http.Server, ctx context.Context, wg *sync.WaitGroup, objstr string) {
				defer wg.Done()
				if err := s.Shutdown(ctx); err != nil {
					TLOG.Errorf("grace shutdown http %s failed within %v, err%v", objstr, graceShutdownTimeout, err)
				} else {
					TLOG.Infof("grace shutdown http %s success %d", objstr, pid)
				}
			}(s, ctx, &wg, obj)
		}

		if s, ok := a.goSvrs[obj]; ok {
			wg.Add(1)
			go func(s *transport.TarsServer, ctx context.Context, wg *sync.WaitGroup, objstr string) {
				defer wg.Done()
				if err := s.Shutdown(ctx); err != nil {
					TLOG.Errorf("grace shutdown tars %s failed within %v, err: %v", objstr, graceShutdownTimeout, err)
				} else {
					TLOG.Infof("grace shutdown tars %s success %d", objstr, pid)
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
		TLOG.Infof("grace shutdown all success within : %v", graceShutdownTimeout)
	case <-time.After(graceShutdownTimeout):
		TLOG.Errorf("grace shutdown timeout within : %v", graceShutdownTimeout)
	}

	a.teerDown(nil)
}

func (a *application) teerDown(err error) {
	a.shutdownOnce.Do(func() {
		if err != nil {
			ReportNotifyInfo(NotifyNormal, "server is fatal: "+err.Error())
			fmt.Println(err)
			TLOG.Error(err)
		}
		a.shutdown <- true
	})
}

func (a *application) handleSignal() {
	usrFun, killFunc := a.graceRestart, a.graceShutdown
	grace.GraceHandler(usrFun, killFunc)
}

func (a *application) mainLoop() {
	ha := new(NodeFHelper)
	comm := a.Communicator()
	node := a.svrCfg.Node
	app := a.svrCfg.App
	server := a.svrCfg.Server
	container := a.svrCfg.Container
	ha.SetNodeInfo(comm, node, app, server, container)

	svrCfg := a.ServerConfig()
	go ha.ReportVersion(svrCfg.Version)
	go ha.KeepAlive("") //first start
	go a.handleSignal()
	// registrar service
	ctx := context.Background()
	go a.registryAdapters(ctx)

	loop := time.NewTicker(svrCfg.MainLoopTicker)
	for {
		select {
		case <-a.shutdown:
			ReportNotifyInfo(NotifyNormal, "stop")
			return
		case <-loop.C:
			if atomic.LoadInt32(&a.isShutdowning) == 1 {
				continue
			}
			for name, adapter := range a.svrCfg.Adapters {
				if adapter.Protocol == "not_tars" {
					// TODO not_tars support
					ha.KeepAlive(name)
					continue
				}
				if s, ok := a.goSvrs[adapter.Obj]; ok {
					if !s.IsZombie(svrCfg.ZombieTimeout) {
						ha.KeepAlive(name)
					}
				}
			}
		}
	}
}
