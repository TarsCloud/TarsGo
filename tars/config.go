package tars

import (
	"github.com/TarsCloud/TarsGo/tars/util/endpoint"
)

var svrCfg *serverConfig
var cltCfg *clientConfig

func GetServerConfig() *serverConfig {
	Init() //引用配置前先初始化应用
	return svrCfg
}

func GetClientConfig() *clientConfig {
	Init() //引用配置前先初始化应用
	return cltCfg
}

type adapterConfig struct {
	Endpoint endpoint.Endpoint
	Protocol string
	Obj      string
	Threads  int
}

type serverConfig struct {
	Node      string
	App       string
	Server    string
	LogPath   string
	LogSize   string
	LogLevel  string
	Version   string
	LocalIP   string
	BasePath  string
	DataPath  string
	config    string
	notify    string
	log       string
	netThread int
	Adapters  map[string]adapterConfig

	Container   string
	Isdocker    bool
	Enableset   bool
	Setdivision string
}

type clientConfig struct {
	Locator                 string
	stat                    string
	property                string
	modulename              string
	refreshEndpointInterval int
	reportInterval          int
	AsyncInvokeTimeout      int
}
