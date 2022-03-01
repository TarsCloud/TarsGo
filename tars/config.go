package tars

import (
	"github.com/TarsCloud/TarsGo/tars/util/tools"
	"time"

	"github.com/TarsCloud/TarsGo/tars/util/endpoint"
)

var svrCfg *serverConfig
var cltCfg *clientConfig

// GetServerConfig Get server config
func GetServerConfig() *serverConfig {
	Init()
	return svrCfg
}

// GetClientConfig Get client config
func GetClientConfig() *clientConfig {
	Init()
	return cltCfg
}

type adapterConfig struct {
	Endpoint endpoint.Endpoint
	Protocol string
	Obj      string
	Threads  int
}

type serverConfig struct {
	Node     string
	App      string
	Server   string
	LogPath  string
	LogSize  uint64
	LogNum   uint64
	LogLevel string
	Version  string
	LocalIP  string
	Local    string
	BasePath string
	DataPath string
	Config   string
	Notify   string
	Log      string
	Adapters map[string]adapterConfig

	Container   string
	Isdocker    bool
	Enableset   bool
	Setdivision string
	//add server timeout
	AcceptTimeout time.Duration
	ReadTimeout   time.Duration
	WriteTimeout  time.Duration
	HandleTimeout time.Duration
	IdleTimeout   time.Duration
	ZombieTimeout time.Duration
	QueueCap      int
	//add tcp config
	TCPReadBuffer  int
	TCPWriteBuffer int
	TCPNoDelay     bool
	//add routine number
	MaxInvoke int32
	//add adapter & report config
	PropertyReportInterval  time.Duration
	StatReportInterval      time.Duration
	MainLoopTicker          time.Duration
	StatReportChannelBufLen int32
	MaxPackageLength        int
	GracedownTimeout        time.Duration

	// tls
	CA           string
	Cert         string
	Key          string
	VerifyClient bool
	Ciphers      string
}

type clientConfig struct {
	Locator                 string
	Stat                    string
	Property                string
	ModuleName              string
	RefreshEndpointInterval int
	ReportInterval          int
	CheckStatusInterval     int
	KeepAliveInterval       int
	AsyncInvokeTimeout      int
	//add client timeout
	ClientQueueLen         int
	ClientIdleTimeout      time.Duration
	ClientReadTimeout      time.Duration
	ClientWriteTimeout     time.Duration
	ClientDialTimeout      time.Duration
	ReqDefaultTimeout      int32
	ObjQueueMax            int32
	AdapterProxyTicker     time.Duration
	AdapterProxyResetCount int
}

func newServerConfig() *serverConfig {
	return &serverConfig{
		Node:                    "",
		App:                     "",
		Server:                  "",
		LogPath:                 "",
		LogSize:                 defaultRotateSizeMB,
		LogNum:                  defaultRotateN,
		LogLevel:                "INFO",
		Version:                 TarsVersion,
		LocalIP:                 tools.GetLocalIP(),
		Local:                   "",
		BasePath:                "",
		DataPath:                "",
		Config:                  "",
		Notify:                  "",
		Log:                     "",
		Adapters:                make(map[string]adapterConfig),
		Container:               "",
		Isdocker:                false,
		Enableset:               false,
		Setdivision:             "",
		AcceptTimeout:           tools.ParseTimeOut(AcceptTimeout),
		ReadTimeout:             tools.ParseTimeOut(ReadTimeout),
		WriteTimeout:            tools.ParseTimeOut(ReadTimeout),
		HandleTimeout:           tools.ParseTimeOut(HandleTimeout),
		IdleTimeout:             tools.ParseTimeOut(IdleTimeout),
		ZombieTimeout:           tools.ParseTimeOut(ZombieTimeout),
		QueueCap:                QueueCap,
		TCPReadBuffer:           TCPReadBuffer,
		TCPWriteBuffer:          TCPWriteBuffer,
		TCPNoDelay:              TCPNoDelay,
		MaxInvoke:               MaxInvoke,
		PropertyReportInterval:  tools.ParseTimeOut(PropertyReportInterval),
		StatReportInterval:      tools.ParseTimeOut(StatReportInterval),
		MainLoopTicker:          tools.ParseTimeOut(MainLoopTicker),
		StatReportChannelBufLen: StatReportChannelBufLen,
		MaxPackageLength:        MaxPackageLength,
		GracedownTimeout:        tools.ParseTimeOut(GracedownTimeout),
	}
}

func newClientConfig() *clientConfig {
	conf := &clientConfig{
		Stat:                    Stat,
		Property:                Property,
		ModuleName:              ModuleName,
		RefreshEndpointInterval: refreshEndpointInterval,
		ReportInterval:          reportInterval,
		CheckStatusInterval:     checkStatusInterval,
		KeepAliveInterval:       keepAliveInverval,
		AsyncInvokeTimeout:      AsyncInvokeTimeout,
		ClientQueueLen:          ClientQueueLen,
		ClientIdleTimeout:       tools.ParseTimeOut(ClientIdleTimeout),
		ClientReadTimeout:       tools.ParseTimeOut(ClientReadTimeout),
		ClientWriteTimeout:      tools.ParseTimeOut(ClientWriteTimeout),
		ClientDialTimeout:       tools.ParseTimeOut(ClientDialTimeout),
		ReqDefaultTimeout:       ReqDefaultTimeout,
		ObjQueueMax:             ObjQueueMax,
		AdapterProxyTicker:      tools.ParseTimeOut(AdapterProxyTicker),
		AdapterProxyResetCount:  AdapterProxyResetCount,
	}
	return conf
}
