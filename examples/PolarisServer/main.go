package main

import (
	"fmt"
	"log"
	"os"

	"polarisserver/tars-protocol/TestApp"

	pr "github.com/TarsCloud/TarsGo/contrib/registry/polaris"
	"github.com/TarsCloud/TarsGo/tars"
	"github.com/polarismesh/polaris-go"
)

func main() {
	// Get server config
	cfg := tars.GetServerConfig()

	provider, err := polaris.NewProviderAPI()
	// 或者使用以下方法,则不需要创建配置文件
	// provider, err = api.NewProviderAPIByAddress("127.0.0.1:8091")

	if err != nil {
		log.Fatalf("fail to create providerAPI, err is %v", err)
	}
	defer provider.Destroy()
	// New servant imp
	imp := new(HelloObjImp)
	err = imp.Init()
	if err != nil {
		fmt.Printf("HelloObjImp init fail, err:(%s)\n", err)
		os.Exit(-1)
	}
	// New servant
	app := new(TestApp.HelloObj)
	// Register Servant
	app.AddServantWithContext(imp, cfg.App+"."+cfg.Server+".HelloObj")

	// Run application
	tars.Run(tars.WithRegistry(pr.New(provider, pr.WithNamespace("tars"))))
}
