package main

import (
	"TarPushServer/server/Impl"
	"flag"
	"fmt"
	"github.com/TarsCloud/TarsGo/tars"
)

func main() {
	SrvConfig := "" // tars服务配置

	flag.StringVar(&SrvConfig, "config", "", "path to server config file")
	flag.Parse()

	fmt.Println("srv_conf: ", SrvConfig)

	// 这里赋值为了避免tars解析命令行参数导致二次解析报错
	tars.ServerConfigPath = SrvConfig

	cfg := tars.GetServerConfig()
	imp := &Impl.DemoImp{}
	Impl.GetApp().AddServantWithContext(imp, cfg.App+"."+cfg.Server+".DemoObj")

	//klog.Init(cfg.App, cfg.Server, cfg.LogPath, int(cfg.LogNum))

	//klog.CONSOLE.Info("server running...")
	tars.Run()
}
