package main

import (
	"fmt"
	"log"

	"github.com/polarismesh/polaris-go"

	pr "github.com/TarsCloud/TarsGo/contrib/registry/polaris"
	"github.com/TarsCloud/TarsGo/tars"

	"polarisserver/tars-protocol/TestApp"
)

func main() {
	//provider, err := polaris.NewProviderAPI()
	// 或者使用以下方法,则不需要创建配置文件
	provider, err := polaris.NewProviderAPIByAddress("127.0.0.1:8091")
	if err != nil {
		log.Fatalf("fail to create providerAPI, err is %v", err)
	}
	defer provider.Destroy()
	// 注册中心
	comm := tars.NewCommunicator(tars.WithRegistry(pr.New(provider, pr.WithNamespace("tars"))))
	obj := fmt.Sprintf("TestApp.PolarisServer.HelloObj")
	app := new(TestApp.HelloObj)
	comm.StringToProxy(obj, app)
	var out, i int32
	i = 123
	ret, err := app.Add(i, i*2, &out)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(ret, out)
}
