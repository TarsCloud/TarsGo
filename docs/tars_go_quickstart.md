# Tars Go快速指南

## 环境搭建

tars基础组件安装参考[部署](https://github.com/TarsCloud/Tars/tree/master/deploy)

Golang环境准备，tarsgo要求golang版本在1.14.x及以上。

安装TarsGo项目创建脚手架：
```shell
# < go 1.17
go get -u github.com/TarsCloud/TarsGo/tars/tools/tarsgo
# >= go 1.17
go install github.com/TarsCloud/TarsGo/tars/tools/tarsgo@latest
```

安装编译tars协议转Golang工具：

```shell
# < go 1.17
go get -u github.com/TarsCloud/TarsGo/tars/tools/tars2go
# >= go 1.17
go install github.com/TarsCloud/TarsGo/tars/tools/tars2go@latest
```

检查下GOPATH路径下tars是否安装成功。

## 服务命名

使用Tars框架的服务，服务名称由三个部分组成：

- APP： 应用名，标识一组服务的一个小集合，在Tars系统中，应用名必须唯一。例如：TestApp；

- Server： 服务名，提供服务的进程名称，Server名字根据业务服务功能命名，一般命名为：XXServer，例如HelloServer；

- Servant：服务者，提供具体服务的接口或实例。例如:HelloImp；

说明：

一个Server可以包含多个Servant，系统会使用服务的App + Server + Servant，进行组合，来定义服务在系统中的路由名称，称为路由Obj，其名称在整个系统中必须是唯一的，以便在对外服务时，能唯一标识自身。

因此在定义APP时，需要注意APP的唯一性。

例如：TestApp.HelloServer.HelloObj。



##  Tars管理系统

用户登录成功后，会进入Tars管理系统，如下图

![tars_manager_main](../docs/images/tars_web_index.png)

TARS管理系统的菜单树下，有以下功能：

- 业务管理：包括已部署的服务，以及服务管理、发布管理、服务配置、服务监控、特性监控等；
- 运维管理：包括服务部署、扩容、模版管理等；

## 服务部署

服务部署，其实也可以在服务开发后进行，不过建议先做。

如下图：

![new_project](../docs/images/tars_go_quickstart_bushu1.png)

- 应用 ： 服务程序归在哪一个应用下，例如：TestApp。
- 服务名称： 服务程序的标识名字，例如：HelloGo。
- 服务类型：服务程序用什么语言写的，例如：go的选择tars_go。
- 模版：服务程序在启动时，设置的配置文件的名称，默认用tars.default即可。
- 节点： 指服务部署的机器IP。
- Set分组：指设置服务的Set分组信息，Set信息包括3部分：Set名、Set地区、Set组名。
- OBJ名称： 指Servant的名称。
- OBJ绑定IP： 指服务绑定的机器IP，一般与节点一样。
- 端口： OBJ要绑定的端口。
- 端口类型：使用TCP还是UDP。
- 协议： 应用层使用的通信协议，Tars框架默认使用tars协议。
- 线程数： 业务处理线程的数目。
- 最大连接数： 支持的最大连接数。
- 队列最大长度： 请求接收队列的大小。
- 队列超时时间：请求接收队列的超时时间。

点击“提交“，成功后，菜单数下的TestApp应用将出现HelloServer名称，同时将在右侧看到你新增的服务程序信息，如下图：

![service_inactive](../docs/images/tars_go_quickstart_service_inactive.png)

## 服务编写

### 创建服务

运行tarsgo脚手架，自动创建服务必须的文件。

```shell
tarsgo make [App] [Server] [Servant] [GoModuleName]
例如： 
tarsgo make TestApp HelloGo SayHello github.com/Tars/test
```

命令执行后将生成代码至GOPATH中，并以`APP/Server`命名目录，生成代码中也有提示具体路径。

```shell
[root@1-1-1-1 ~]# tarsgo make TestApp HelloGo SayHello github.com/Tars/test
🚀 Creating server TestApp.HelloGo, layout repo is https://github.com/TarsCloud/TarsGo.git, please wait a moment.

已经是最新的。

go: creating new go.mod: module github.com/Tars/test
go: to add module requirements and sums:
	go mod tidy

CREATED HelloGo/SayHello.tars (171 bytes)
CREATED HelloGo/SayHello_imp.go (620 bytes)
CREATED HelloGo/client/client.go (444 bytes)
CREATED HelloGo/config.conf (967 bytes)
CREATED HelloGo/debugtool/dumpstack.go (412 bytes)
CREATED HelloGo/go.mod (37 bytes)
CREATED HelloGo/main.go (517 bytes)
CREATED HelloGo/makefile (193 bytes)
CREATED HelloGo/scripts/makefile.tars.gomod (4181 bytes)
CREATED HelloGo/start.sh (56 bytes)

>>> Great！Done! You can jump in HelloGo
>>> Tips: After editing the Tars file, execute the following cmd to automatically generate golang files.
>>>       /root/gocode/bin/tars2go *.tars
$ cd HelloGo
$ ./start.sh
🤝 Thanks for using TarsGo
📚 Tutorial: https://tarscloud.github.io/TarsDocs/
```

### 定义接口文件

接口文件定义请求方法以及参数字段类型等，有关接口定义文件说明参考tars_tup.md

为了测试我们定义一个echoHello的接口，客户端请求参数是短字符串如 "tars"，服务响应"hello tars".

```shell
# cat HelloGo/SayHello.tars 
module TestApp{
    interface SayHello{
        int echoHello(string name, out string greeting); 
    };
};
```

**注意**： 参数中**out**修饰关键字标识输出参数。

### 服务端开发

首先把tars协议文件转化为Golang语言形式

```shell
tars2go  -outdir=tars-protocol -module=github.com/Tars/test SayHello.tars
```

现在开始实现服务端的逻辑：客户端传来一个名字，服务端回应hello name。

```shell
cat HelloGo/SayHello_imp.go
```

```go
package main
import "context"
type SayHelloImp struct {
}

func (imp *SayHelloImp) EchoHello(ctx context.Context, name string, greeting *string) (int32, error) {
    *greeting = "hello " + name
    return 0, nil
}
```

**注意**： 这里函数名要大写，Go语言方法导出规定。

编译main函数，初始代码以及有tars框架实现了。

cat  HelloGo/main.go

```go
package main

import (
    "github.com/TarsCloud/TarsGo/tars"
    
    "github.com/Tars/test/tars-protocol/TestApp"
)

func main() {
    // Get server config
    cfg := tars.GetServerConfig()
  
    // New servant imp
    imp := new(SayHelloImp)
    // New servant
    app := new(TestApp.SayHello)
    // Register Servant
    app.AddServantWithContext(imp, cfg.App+"."+cfg.Server+".SayHelloObj")
  
    // Run application
    tars.Run()
}
```

编译生成可执行文件，并打包发布包。

```shell
cd HelloGo && make && make tar
```

将生成可执行文件HelloGo和发布包HelloGo.tgz

### 客户端开发

```go
package main

import (
    "fmt"
  
    "github.com/TarsCloud/TarsGo/tars"
  
    "github.com/Tars/test/tars-protocol/TestApp"
)

//只需初始化一次，全局的
var comm *tars.Communicator
func main() {
    comm = tars.NewCommunicator()
    obj := "TestApp.HelloGo.SayHelloObj@tcp -h 127.0.0.1 -p 10015 -t 60000"
    app := new(TestApp.SayHello)
    /*
       // if your service has been registered at tars registry
       obj := "TestApp.HelloGo.SayHelloObj"
       // tarsregistry service at 192.168.1.1:17890
       comm.SetProperty("locator", "tars.tarsregistry.QueryObj@tcp -h 192.168.1.1 -p 17890")
    */
  
    comm.StringToProxy(obj, app)
    reqStr := "tars"
    var resp string
    ret, err := app.EchoHello(reqStr, &resp)
    if err != nil {
      fmt.Println(err)
      return
    }
    fmt.Println("ret: ", ret, "resp: ", resp)
}
```

- TestApp依赖是tars2go生成的代码。

- obj指定服务端地址端口，如果服务端未在主控注册，则需要知道服务端的地址和端口并在Obj中指定，在例子中，协议为TCP，服务端地址为本地地址，端口为3002。如果有多个服务端，则可以这样写`TestApp.HelloGo.SayHelloObj@tcp -h 127.0.0.1 -p 9985:tcp -h 192.168.1.1 -p 9983`这样请求可以分散到多个节点。

  如果已经在主控注册了服务，则不需要写死服务端地址和端口，但在初始化通信器时需要指定主控的地址。

- com通信器，用于与服务端通信。

编译测试

```shell
# go build ./client/client.go
# ./client/client
ret:  0 resp:  hello tars 
```



### HTTP 服务开发

tarsgo支持http服务，按照上面的步骤创建好服务，tarsgo中处理http请求是在GO原生中的封装，所以使用很简单。

```go
package main

import (
    "net/http"
    "github.com/TarsCloud/TarsGo/tars"
)

func main() {
    mux := &tars.TarsHttpMux{}
    mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
      w.Write([]byte("Hello tars"))
    })
    cfg := tars.GetServerConfig()
    tars.AddHttpServant(mux, cfg.App+"."+cfg.Server+".HttpSayHelloObj") //Register http server
    tars.Run()
}
```

另外还可以直接调用其他tars服务，调用方式和“客户端开发”提到一样。

## 服务发布

在管理系统的菜单树下，找到你部署的服务，点击进入服务页面。

选择“发布管理”，选中要发布的节点，点击“发布选中节点”，点击“上传发布包”，选择已经编译好的发布包，如下图：

![release](../docs/images/tars_go_quickstart_release.png)

上传好发布包后，点击“选择发布版本”下拉框就会出现你上传的服务程序，选择最上面的一个（最新上传的）。

点击“发布”，服务开始发布，发布成功后，出现下面的界面，如下图：

![service_ok](../docs/images/tars_go_quickstart_service_ok.png)


