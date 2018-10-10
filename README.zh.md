# Tarsgo  文档

## 关于
- 	Tarsgo是基于Golang编程语言使用Tars协议的高性能RPC框架。随着docker,k8s,etcd等容器化技术的兴起，Go语言变得流行起来。Go的goroutine并发机制使Go非常适合用于大规模高并发后端服务程序的开发。 Go语言具有接近C/C++的性能和接近python的生产力。在腾讯，一部分现有的C++开发人员正逐渐向Go转型，Tars作为广泛使用的RPC框架，现已支持C++/Java/Nodejs/Php，其与Go语言的结合已成为大势所趋。因此，在广大用户的呼声中我们推出了Tarsgo,并且已经将它应用于腾讯地图、应用宝、互联网+以及其他项目中。
- 关于tars的整体架构和设计理念，请阅读 [Tars介绍](https://github.com/TarsCloud/Tars/blob/master/Introduction.md)

## 功能特性
- Tars2go工具: tars文件自动生成并转换为go语言，包含用go语言实现的RPC服务端/客户端代码
- go语言版本的tars的序列化和反序列化包
- 服务端支持心跳上报，统计监控上报，自定义命令处理，基础日志
- 客户端支持直接连接和路由访问，自动重新连接，定期刷新节点状态以及支持UDP/TCP协议
- 提供远程日志
- 提供特性监控上报
- 提供set分组
- 提供 protocol buffers 支持， 详见 [pb2tarsgo](tars/tools/pb2tarsgo/README.md) 



## 安装
- 对于安装OSS和其他基本服务, 请[安装文档](https://github.com/TarsCloud/Tars/blob/master/Install.zh.md)，
快速安装，请查看[快速部署](https://github.com/TarsCloud/Tars/tree/master/deploy)
- 要求Go 1.9.x 或以上版本,请查看https://golang.org/doc/install
- go get -u github.com/TarsCloud/TarsGo/tars


## 快速开始
- 快速开始，请查看 [tars\_go\_quickstart.md](docs/tars_go_quickstart.md)

## 性能数据
- 查看[性能数据](docs/tars_go_performance.md)


## 使用
### 1 服务端
 - 下面是一个完整的示例，用于说明如何使用tarsgo去构建服务端。
  
#### 1.1 接口定义

在 $GOPATH/src下编写一个tars文件，如hello.tars , 比如 $GOPATH/src/TestApp/TestServer/hello.tars.
有关tars协议的更多详细信息, 请查看 https://github.com/TarsCloud/TarsTup/blob/master/docs-en/tars_tup.md
	
```
	
	module TestApp
	{
	
	interface Hello
	{
	    int test();
	    int testHello(string sReq, out string sRsp);
	};
	
	}; 
	
```
	

#### 1.2 编译接口定义文件

##### 1.2.1 构建 tars2go
编译tars2go工具并将tars2go二进制文件复制到$PATH

	cd $GOPATH/src/github.com/TarsCloud/TarsGo/tars/tools/tars2go && go build . 
    cp tarsgo $GOPTAH/bin 

##### 1.2.2 编译tars文件并转成go文
	tars2go --outdir=./vendor hello.tars
#### 1.3 接口实现
```
package main

import (
    "github.com/TarsCloud/TarsGo/tars"

    "TestApp"
)

type HelloImp struct {
}

//implete the Test interface
func (imp *HelloImp) Test() (int32, error) {
    return 0, nil 
}

//implete the testHello interface

func (imp *HelloImp) TestHello(in string, out *string) (int32, error) {
    *out = in
    return 0, nil 
}


func main() { //Init servant
    imp := new(HelloImp)                                    //New Imp
    app := new(TestApp.Hello)                               //New init the A Tars
    cfg := tars.GetServerConfig()                           //Get Config File Object
    app.AddServant(imp, cfg.App+"."+cfg.Server+".HelloObj") //Register Servant
    tars.Run()
}


```

说明:

- HelloImp是结构体，你在里面实现Hello和Test接口, 注意Test和Hello必须以大写字母开头才能被导出,这是唯一与tars文件定义有所不同的地方。
- TestApp.Hello是由tar2go工具生成的，它可以在./vendor/TestApp/Hello_IF.go中找到，其中包含一个名为TestApp的软件包，它与tars文件的TestApp模块一样。
-  tars.GetServerConfig()用于获得服务端配置。
-  cfg.App+"."+cfg.Server+".HelloObj" 是绑定到Servant的对象名，客户端将使用此名称访问服务端.



#### 1.4 服务端配置

tars.GetServerConfig()返回服务端配置，其定义如下:

```
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


```

- Node: 本地tarsnode地址，只有你使用tars平台部署才会使用这个参数.
- APP: 应用名.
- Server: 服务名.
- LogPath: 保存日志的目录.
- LogSize: 轮换日志的大小.
- LogLevel: 轮换日志的级别.
- Version: Tarsg的版本.
- LocalIP: 本地ip地址.
- BasePath: 二进制文件的基本路径.
- DataPath: 一些缓存文件存储路径.
- config: 获取配置的配置中心，如tars.tarsconfig.ConfigObj
- notify： 上报通知报告的通知中心，如tars.tarsnotify.NotifyObj
- log： 远程日志中心，如tars.tarslog.LogObj
- netThread: 保留用于控制接收和发送包的go线程.
- Adapters:  每个adapter适配器的指定配置.
- Contianer: 保留供以后使用，用于存储容器名称.
- Isdocker: 保留供以后使用，用于指定服务是否在容器内运行.
- Enableset: 如果使用了set，则为True.
- Setdivision: 指定哪个set，如gray.sz.*

如下是一个服务端配置的例子:
```
<tars>
  <application>
      enableset=Y
      setdivision=gray.sz.*
    <server>
       node=tars.tarsnode.ServerObj@tcp -h 10.120.129.226 -p 19386 -t 60000
       app=TestApp
       server=HelloServer
       localip=10.120.129.226
       local=tcp -h 127.0.0.1 -p 20001 -t 3000
       basepath=/usr/local/app/tars/tarsnode/data/TestApp.HelloServer/bin/
       datapath=/usr/local/app/tars/tarsnode/data/TestApp.HelloServer/data/
       logpath=/usr/local/app/tars/app_log/
       logsize=10M
       config=tars.tarsconfig.ConfigObj
       notify=tars.tarsnotify.NotifyObj
       log=tars.tarslog.LogObj
       #timeout for deactiving , ms.
       deactivating-timeout=2000
       logLevel=DEBUG
    </server>
  </application>
</tars>

```

#### 1.5 适配器
适配器为每个对象绑定ip和端口.在服务端代码实现的例子中， 
app.AddServant(imp, cfg.App+"."+cfg.Server+".HelloObj")完成HelloObj的适配器配置和实现的绑定。适配器的完整例子如下：

```
<tars>
  <application>
    <server>
       #each adapter configuration 
       <TestApp.HelloServer.HelloObjAdapter>
            #allow Ip for white list.
            allow
            # ip and port to listen on  
            endpoint=tcp -h 10.120.129.226 -p 20001 -t 60000
            #handlegroup
            handlegroup=TestApp.HelloServer.HelloObjAdapter
            #max connection 
            maxconns=200000
            #portocol, only tars for now.
            protocol=tars
            #max capbility in handle queue.
            queuecap=10000
            #timeout in ms for the request in the queue.
            queuetimeout=60000
            #servant 
            servant=TestApp.HelloServer.HelloObj
            #threads in handle server side implement code. goroutine for golang.
            threads=5
       </TestApp.HelloServer.HelloObjAdapter>
    </server>
  </application>
</tars>
```


#### 1.6 服务端启动 

如下命令用于启动服务端：
```
./HelloServer --config=config.conf
```
请参阅下面的config.conf的完整示例，稍后我们将解释客户端配置。


```
<tars>
  <application>
    enableset=n
    setdivision=NULL
    <server>
       node=tars.tarsnode.ServerObj@tcp -h 10.120.129.226 -p 19386 -t 60000
       app=TestApp
       server=HelloServer
       localip=10.120.129.226
       local=tcp -h 127.0.0.1 -p 20001 -t 3000
       basepath=/usr/local/app/tars/tarsnode/data/TestApp.HelloServer/bin/
       datapath=/usr/local/app/tars/tarsnode/data/TestApp.HelloServer/data/
       logpath=/usr/local/app/tars/app_log/
       logsize=10M
       config=tars.tarsconfig.ConfigObj
       notify=tars.tarsnotify.NotifyObj
       log=tars.tarslog.LogObj
       deactivating-timeout=2000
       logLevel=DEBUG
       <TestApp.HelloServer.HelloObjAdapter>
            allow
            endpoint=tcp -h 10.120.129.226 -p 20001 -t 60000
            handlegroup=TestApp.HelloServer.HelloObjAdapter
            maxconns=200000
            protocol=tars
            queuecap=10000
            queuetimeout=60000
            servant=TestApp.HelloServer.HelloObj
            threads=5
       </TestApp.HelloServer.HelloObjAdapter>
    </server>
    <client>
       locator=tars.tarsregistry.QueryObj@tcp -h 10.120.129.226 -p 17890
       sync-invoke-timeout=3000
       async-invoke-timeout=5000
       refresh-endpoint-interval=60000
       report-interval=60000
       sample-rate=100000
       max-sample-count=50
       asyncthread=3
       modulename=TestApp.HelloServer
    </client>
  </application>
</tars>
```


### 2 客户端
用户可以轻松编写客户端代码，而无需编写任何指定协议的通信代码.
#### 2.1 客户端例子
请参阅下面的一个客户端例子:

```

package main

import (
    "fmt"
    "github.com/TarsCloud/TarsGo/tars"
    "TestApp"
)
//tars.Communicator should only init once and be global
var comm *tars.Communicator

func main() {
    comm = tars.NewCommunicator()
    obj := "TestApp.TestServer.HelloObj@tcp -h 127.0.0.1 -p 10015 -t 60000"
    app := new(TestApp.Hello)
    comm.StringToProxy(obj, app)
	var req string="Hello Wold"
    var res string
    ret, err := app.TestHello(req, &out)
    if err != nil {
        fmt.Println(err)
        return
    }   
    fmt.Println(ret, out)

```

说明:

- TestApp包是由tars2go工具使用tars协议文件生成的.
- comm: Communicator用于与服务端进行通信，它应该只初始化一次并且是全局的.
- obj: 对象名称，用于指定服务端的ip和端口。通常在"@"符号之前我们只需要对象名称.
- app: 与tars文件中的接口关联的应用程序。 在本例中它是TestApp.Hello.
- StringToProxy: StringToProxy方法用于绑定对象名称和应用程序，如果不这样做，通信器将不知道谁与应用程序通信 .
- req, res: 在tars文件中定义的输入和输出参数,用于在TestHello方法中.
- app.TestHello用于调用tars文件中定义的方法，并返回ret和err.

#### 2.2 通信器
通信器是为客户端发送和接收包的一组资源，其最终管理每个对象的socket通信。在一个程序中你只需要一个通信器。

```
var comm *tars.Communicato
comm = tars.NewCommunicator()
comm.SetProperty("property", "tars.tarsproperty.PropertyObj")
comm.SetProperty("locator", "tars.tarsregistry.QueryObj@tcp -h ... -p ...")
```

描述:
> * 通信器配置文件的格式将在后面描述.
> * 可以在没有配置文件的情况下配置通信器，并且所有参数都具有默认值.
> * 通信器也可以通过“SetProperty”方法直接初始化.
> * 如果您不需要配置文件，则必须自己设置locator参数.

通信器属性描述:
> * locator:主控服务的地址必须采用“ip port”格式。 如果你不需要主控来查找服务，则无需配置此项.
> * important  async-invoke-timeout:客户端调用的最大超时时间（以毫秒为单位），此配置的默认值为3000.
> * sync-invoke-timeout:现在没用于tarsgo.
> * refresh-endpoint-interval:定期访问主控以获取信息的时间间隔（以毫秒为单位），此配置的默认值为一分钟.
> * stat:在模块之间调用的服务的地址。 如果未配置此项，则表示将直接丢弃上报的数据.
> * property:服务上报其属性的地址。 如果未配置，则表示将直接丢弃上报的数据.
> * report-interval:现在没用于tarsgo.
> * asyncthread: 已被tarsgo舍弃.
> * modulename: 模块名称，默认值是可执行程序的名称。

通信器配置文件的格式如下：
```
<tars>
  <application>
    #The configuration required by the proxy
    <client>
        #address
        locator                     = tars.tarsregistry.QueryObj@tcp -h 127.0.0.1 -p 17890
        #The maximum timeout (in milliseconds) for synchronous calls.
        sync-invoke-timeout         = 3000
        #The maximum timeout (in milliseconds) for asynchronous calls.
        async-invoke-timeout        = 5000
        #The maximum timeout (in milliseconds) for synchronous calls.
        refresh-endpoint-interval   = 60000
        #Used for inter-module calls
        stat                        = tars.tarsstat.StatObj
        #Address used for attribute reporting
        property                    = tars.tarsproperty.PropertyObj
        #report time interval
        report-interval             = 60000
        #The number of threads that process asynchronous responses
        asyncthread                 = 3
        #The module name
        modulename                  = Test.HelloServer
    </client>
  </application>
</tars>
```
#### 2.3 超时控制
如果你想在客户端使用超时控制，请使用以ms为单位的TarsSetTimeout。
```
    app := new(TestApp.Hello)
    comm.StringToProxy(obj, app)
    app.TarsSetTimeout(3000)
```

#### 2.4 接口调用

本节详细介绍了Tars客户端如何远程调用服务端。

首先，简要描述Tars客户端的寻址模式。 其次，它将介绍客户端的调用方法，包括但不限于单向调用，同步调用，异步调用，hash调用等。

##### 2.4.1 寻址模式简介. 

Tars服务的寻址模式通常可以分为两种方式：服务名称在master上注册了，服务名称未在master上注册。 master是专用于注册服务节点信息的名字服务（路由服务）。

把服务名添加到名字服务中是通过操作管理平台实现。

对于未在master中注册的服务，可以将其分类为直接寻址，即在调用服务之前需要指定服务提供者的IP地址。 客户端需要在调用服务时指定HelloObj对象的特定地址，即Test.HelloServer.HelloObj@tcp -h 127.0.0.1 -p 9985

Test.HelloServer.HelloObj: 对象名

tcp:Tcp协议

-h:指定主机地址,这里是127.0.0.1

-p:端口,这里是9985

如果HelloServer在两台服务器上运行，则应用程序初始化如下:
```
    obj:= "Test.HelloServer.HelloObj@tcp -h 127.0.0.1 -p 9985:tcp -h 192.168.1.1 -p 9983"
    app := new(TestApp.Hello)
    comm.StringToProxy(obj, app)
```
HelloObj的地址设置为两个服务器的地址。 此时，请求将被分发到两个服务器（可以指定分发方法，这里不再介绍）。 如果一台服务器关闭，请求将自动分配给另一台服务器，服务器将定期重新启动。

对于在master中注册的服务，将根据服务名称对服务进行寻址。 当客户端请求服务时，它不需要指定HelloServer的特定地址，但是在生成通信器或初始化通信器时需要指定`registry`的地址。

以下通过设置通信器的参数显示主控的地址：
```
var *tars.Communicator
comm = tars.NewCommunicator()
comm.SetProperty("locator", "tars.tarsregistry.QueryObj@tcp -h ... -p ...")
```
由于客户端需要依赖主控的地址，因此主控还必须具有容错能力。 主控的容错方法与上面相同，即指定了两个主控的地址。
##### 2.4.2. 单向调用
TODO. tarsgo暂未支持.

##### 2.4.3. 同步调用
```
package main

import (
    "fmt"
    "github.com/TarsCloud/TarsGo/tars"
    "TestApp"
)

var *tars.Communicator
func main() {
    comm = tars.NewCommunicator()
    obj := "TestApp.TestServer.HelloObj@tcp -h 127.0.0.1 -p 10015 -t 60000"
    app := new(TestApp.Hello)
    comm.StringToProxy(obj, app)
	var req string="Hello Wold"
    var res string
    ret, err := app.TestHello(req, &out)
    if err != nil {
        fmt.Println(err)
        return
    }   
    fmt.Println(ret, out)

```

##### 2.4.4 异步调用
tarsgo可以使用goroutine轻松使用异步调用。 与cpp不同，我们不需要实现回调函数。

```
package main

import (
    "fmt"
    "github.com/TarsCloud/TarsGo/tars"
    "time"
    "TestApp"
)
var *tars.Communicator
func main() {
    comm = tars.NewCommunicator()
    obj := "TestApp.TestServer.HelloObj@tcp -h 127.0.0.1 -p 10015 -t 60000"
    app := new(TestApp.Hello)
    comm.StringToProxy(obj, app)
	go func(){
		var req string="Hello Wold"
    	var res string
    	ret, err := app.TestHello(req, &out)
    	if err != nil {
        	fmt.Println(err)
        	return
    	} 
		fmt.Println(ret, out)
	}()
    time.Sleep(1)  

```

##### 2.4.5 通过set调用
客户端可以通过set来调用服务端，只需要配置上文提到的配置文件，其中enableset置为y，setdivision比如设置为gray.sz. *。 有关更多详细信息，请参阅https://github.com/TarsCloud/Tars/blob/master/docs-en/tars_idc_set.md。
如果您想手动通过set调用，tarsgo将很快支持此功能。

##### 2.4.6. Hash调用

由于可以部署多个服务端，因此客户端的请求会随机分发到服务端上，但在某些情况下，希望始终将某些请求发送到特定的服务端。 在这种情况下，Tars提供了一种简单的实现方法，称为hash调用。 Tarsgo很快将支持此功能。

### 3   tars定义的返回码.
```
//Define the return code given by the TARS service
const int TARSSERVERSUCCESS       = 0;    //Server-side processing succeeded
const int TARSSERVERDECODEERR     = -1;   //Server-side decoding exception
const int TARSSERVERENCODEERR     = -2;   //Server-side encoding exception
const int TARSSERVERNOFUNCERR     = -3;   //There is no such function on the server side
const int TARSSERVERNOSERVANTERR  = -4;   //The server does not have the Servant object
const int TARSSERVERRESETGRID     = -5;   // server grayscale state is inconsistent
const int TARSSERVERQUEUETIMEOUT  = -6;   //server queue exceeds limit
const int TARSASYNCCALLTIMEOUT    = -7;   // Asynchronous call timeout
const int TARSINVOKETIMEOUT       = -7;   //call timeout
const int TARSPROXYCONNECTERR     = -8;   //proxy link exception
const int TARSSERVEROVERLOAD      = -9;   //Server overload, exceeding queue length
const int TARSADAPTERNULL         = -10;  //The client routing is empty, the service does not exist or all services are down.
const int TARSINVOKEBYINVALIDESET = -11;  //The client calls the set rule illegally
const int TARSCLIENTDECODEERR     = -12;  //Client decoding exception
const int TARSSERVERUNKNOWNERR    = -99;  //The server is in an abnormal position
```


### 4 日志
使用tarsgo轮换日志的快速示例：

```
TLOG := tars.GetLogger("TLOG")
TLOG.Debug("Debug logging")
```
这将创建一个在tars/util/rogger中定义的*Rogger.Logger，并在调用GetLogger之后，会在config.conf中定义的Logpath下创建一个日志文件，其名称为cfg.App + "." + cfg.Server + "_" +名称，该日志文件将在100MB（默认）后轮换，最大轮换文件数为10（默认）。 

如果你不想按文件大小轮换日志。 例如，你想要按天轮换，使用：

```
TLOG := tars.GetDayLogger("TLOG",1)
TLOG.Debug("Debug logging")
```
使用GetHourLogger("TLOG",1)按小时轮换日志。如果你想打日志到config.conf中定义的名为tars.tarslog.LogObj的远程服务器上，你不得不先配置一个日志服务器。可以在tars/protocol/res/LogF.tars中找到完整的tars文件定义，可以在Tencent/Tars/cpp/framework/LogServer中查找日志服务器。快速示例如下：
```
TLOG := GetRemoteLogger("TLOG")
TLOG.Debug("Debug logging")

```
如果你想设置日志等级，你可以在Tencent/Tars/web下的tars项目提供的OSS平台上设置它。
如果你想自定义你的日志，请在tars/util/logger，tars/logger.go 和tars/remotelogger.go中的查看更多细节。
### 5  服务管理

Tars服务框架支持动态接收命令来处理相关的业务逻辑，例如动态更新配置

tarsgo目前有tars.viewversion / tars.setloglevel管理命令。 用户可以从oss发送管理命令来查看版本或设置日志等级。

如果你想定义你自己的管理命令，请看下面的例子：
```
func helloAdmin(who string ) (string, error) {
	return who, nil
}
tars.RegisterAdmin("tars.helloAdmin",  helloAdmin)

```

然后你可以发送自定义的管理命令“tars.helloAdmin tarsgo”，tarsgo将在浏览器中显示。

举例:
```
// A function  should be in this format
type adminFn func(string) (string, error)

//then u should registry this function using

func RegisterAdmin(name string, fn adminFn)
```

### 6 统计上报

上报统计信息是向Tars框架内的tarsstat上报耗时信息和其他信息。 无需用户开发，只需在程序初始化期间正确设置相关信息后，就可以在框架内自动报告（包括客户端和服务端）。

客户端调用上报接口后，会暂时将信息存储在内存中，当到达某个时间点时，会向tarsstat服务上报（默认为1分钟上报一次）。 我们将两个上报时间点之间的时间间隔称为统计间隔，在统计间隔中会执行诸如聚合和比较相同key的一些操作。
示例代码如下：
```
//for error
ReportStat(msg, 0, 1, 0)

//for success
ReportStat(msg, 1, 0, 0)


//func ReportStat(msg *Message, succ int32, timeout int32, exec int32)
//see more detail in tars/statf.go
```

描述:
> * 通常，我们不必关心统计上报，每次客户端调用服务端之后，无论成功与否，tarsgo框架都将会上报。 如果你设置正确，将在Web管理系统中显示成功率，失败率，平均耗时等。
> * 如果主服务部署在Web管理系统上，则无需定义Communicator、设置tarsregistry，tarsstat等的配置，该服务将会自动上报这些信息。
> * 如果未在Web管理系统上部署主服务或程序，则需要定义Communicator，设置tarsregistry，tarsstat等，以便您可以在Web管理系统上查看被调服务的服务监控。
> * 数据定期上报是可以在通信器的配置中设置。

### 7 异常上报
为了更好地监控，TARS框架支持直接向程序中的tarsnotify上报异常情况，并可在WEB管理页面上查看。

该框架提供了三个宏来上报不同类型的异常：
```
tars.reportNotifyInfo("Get data from mysql error!")
```

Info是一个字符串，可以直接将字符串上报给tarsnotify。 上报的字符串可以在页面上看到，随后，我们可以根据上报的信息进行报警。

### 8 特性监控

为了便于业务统计，TARS框架还支持在Web管理平台上显示信息。

目前支持的统计类型包括：
> * Sum(sum) //计算每个上报值的总和
> * Average(avg) //计算每个上报值的均值
> * Distribution(distr) //计算每个上报的分布，其参数是一个列表，来计算每个区间的概率分布
> * Maximum(max) //计算每个上报值的最大值
> * Minimum(min) // 计算每个上报值的最小值
> * Count(count) //计算上报次数

示例代码如下：
```
    sum := tars.NewSum()
    count := tars.NewCount()
    max := tars.NewMax()
    min := tars.NewMin()
    d := []int{10, 20, 30, 50} 
    distr := tars.NewDistr(d)
    p := tars.CreatePropertyReport("testproperty", sum, count, max, min, distr)
    for i := 0; i < 5; i++ {
        v := rand.Intn(100)
        p.Report(v)

    }   

```

描述:
> * 定期上报数据，可以在通信器的配置中设置，目前是每分钟上报一次;
> * 创建一个PropertyReportPtr函数：参数createPropertyReport可以是任何统计方法的集合，示例中使用六种统计方法，通常只需要使用一个或两个;
> * 注意，当你在调用createPropertyReport时，必须在启用服务后创建并保存所创建的对象，然后只需将对象上报，不要在你每次使用时都创建它。

### 9 远程配置
用户可以从OSS设置远程配置。详情请查看https://github.com/TarsCloud/TarsFramework/blob/master/docs-en/tars_config.md . 
如下示例用于说明如何使用此api从远程获取配置文件。

```
import "github.com/TarsCloud/TarsGo/tars"
...
cfg := tars.GetServerConfig()
remoteConf := tars.NewRConf(cfg.App, cfg.Server, cfg.BasePath)
config, _ := remoteConf.GetConfig("test.conf")

...

```

### 10 setting.go
tars包中的setting.go用于控制tarsgo性能和特性。有些选项应该从Getserverconfig()中更新。

```
//number of woker routine to handle client request
//zero means  no contorl ,just one goroutine for a client request.
//runtime.NumCpu() usually best performance in the benchmark.
var MaxInvoke int = 0

const (
	//for now ,some option shuold update from remote config

	//version
	TarsVsersion string = "1.0.0"

	//server

	AcceptTimeout time.Duration = 500 * time.Millisecond
	//zero for not set read deadline for Conn (better  performance)
	ReadTimeout time.Duration = 0 * time.Millisecond
	//zero for not set write deadline for Conn (better performance)
	WriteTimeout time.Duration = 0 * time.Millisecond
	//zero for not set deadline for invoke user interface (better performance)
	HandleTimeout  time.Duration = 0 * time.Millisecond
	IdleTimeout    time.Duration = 600000 * time.Millisecond
	ZombileTimeout time.Duration = time.Second * 10
	QueueCap       int           = 10000000

	//client
	ClientQueueLen     int           = 10000
	ClientIdleTimeout  time.Duration = time.Second * 600
	ClientReadTimeout  time.Duration = time.Millisecond * 100
	ClientWriteTimeout time.Duration = time.Millisecond * 3000
	ReqDefaultTimeout  int32         = 3000
	ObjQueueMax        int32         = 10000

	//report
	PropertyReportInterval time.Duration = 10 * time.Second
	StatReportInterval     time.Duration = 10 * time.Second

	//mainloop
	MainLoopTicker time.Duration = 10 * time.Second

	//adapter
	AdapterProxyTicker     time.Duration = 10 * time.Second
	AdapterProxyResetCount int           = 5

	//communicator default ,update from remote config
	refreshEndpointInterval int = 60000
	reportInterval          int = 10000
	AsyncInvokeTimeout      int = 3000

	//tcp network config
	TCPReadBuffer  = 128 * 1024 * 1024
	TCPWriteBuffer = 128 * 1024 * 1024
	TCPNoDelay     = false
)


```
