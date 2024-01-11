package tars

import (
	"fmt"
	"strings"
	"time"

	"github.com/TarsCloud/TarsGo/tars/util/endpoint"
	"github.com/TarsCloud/TarsGo/tars/util/tools"
)

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
	NodeName string
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
	// add server timeout
	AcceptTimeout time.Duration
	ReadTimeout   time.Duration
	WriteTimeout  time.Duration
	HandleTimeout time.Duration
	IdleTimeout   time.Duration
	ZombieTimeout time.Duration
	QueueCap      int
	// add tcp config
	TCPReadBuffer  int
	TCPWriteBuffer int
	TCPNoDelay     bool
	// add routine number
	MaxInvoke int32
	// add adapter & report config
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

	SampleRate     float64
	SampleType     string
	SampleAddress  string
	SampleEncoding string
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
	// add client timeout
	ClientQueueLen     int
	ClientIdleTimeout  time.Duration
	ClientReadTimeout  time.Duration
	ClientWriteTimeout time.Duration
	ClientDialTimeout  time.Duration
	ReqDefaultTimeout  int32
	ObjQueueMax        int32
	context            map[string]string
}

// GetServerConfig Get server config
func GetServerConfig() *serverConfig {
	return defaultApp.ServerConfig()
}

// GetClientConfig Get client config
func GetClientConfig() *clientConfig {
	return defaultApp.ClientConfig()
}

func (c *clientConfig) ValidateStat() error {
	if c.Stat == "" || (c.LocatorEmpty() && !strings.Contains(c.Stat, "@")) {
		return fmt.Errorf("stat config emptry")
	}
	return nil
}

func (c *clientConfig) ValidateProperty() error {
	if c.Property == "" || (c.LocatorEmpty() && !strings.Contains(c.Property, "@")) {
		return fmt.Errorf("property config emptry")
	}
	return nil
}

func (c *clientConfig) LocatorEmpty() bool {
	return c.Locator == ""
}

// ServerConfig returns server config
func (a *application) ServerConfig() *serverConfig {
	a.init()
	return a.svrCfg
}

// ClientConfig returns client config
func (a *application) ClientConfig() *clientConfig {
	a.init()
	return a.cltCfg
}

func newServerConfig() *serverConfig {
	return &serverConfig{
		LogSize:                 defaultRotateSizeMB,
		LogNum:                  defaultRotateN,
		LogLevel:                "INFO",
		Version:                 Version,
		LocalIP:                 tools.GetLocalIP(),
		Adapters:                make(map[string]adapterConfig),
		Isdocker:                false,
		Enableset:               false,
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
		KeepAliveInterval:       keepAliveInterval,
		AsyncInvokeTimeout:      AsyncInvokeTimeout,
		ClientQueueLen:          ClientQueueLen,
		ClientIdleTimeout:       tools.ParseTimeOut(ClientIdleTimeout),
		ClientReadTimeout:       tools.ParseTimeOut(ClientReadTimeout),
		ClientWriteTimeout:      tools.ParseTimeOut(ClientWriteTimeout),
		ClientDialTimeout:       tools.ParseTimeOut(ClientDialTimeout),
		ReqDefaultTimeout:       ReqDefaultTimeout,
		ObjQueueMax:             ObjQueueMax,
		context:                 make(map[string]string),
	}
	return conf
}

func (c *clientConfig) Context() map[string]string {
	return c.context
}
