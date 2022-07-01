# Usage
- Install protoc

[https://github.com/protocolbuffers/protobuf/releases](https://github.com/protocolbuffers/protobuf/releases)

In the downloads section of each release, you can find pre-built binaries in zip packages: protoc-$VERSION-$PLATFORM.zip. It contains the protoc binary as well as a set of standard .proto files distributed along with protobuf.

- Add tarsrpc plugin for protoc-gen-go
```shell
export PATH=$PATH:$GOPATH/bin
# < go 1.16
go get -u github.com/TarsCloud/TarsGo/tars/tools/pb2tarsgo/protoc-gen-go
# >= go 1.17
go install github.com/TarsCloud/TarsGo/tars/tools/pb2tarsgo/protoc-gen-go@latest
```

# example

- proto file
```proto
syntax = "proto3";
package helloworld;

// The greeting service definition.
service Greeter {
  // Sends a greeting
  rpc SayHello (HelloRequest) returns (HelloReply) {}
}

// The request message containing the user's name.
message HelloRequest {
  string name = 1;
}

// The response message containing the greetings
message HelloReply {
  string message = 1;
}

```

- generate the code
```
protoc --go_out=plugins=tarsrpc:. helloworld.proto
```
- server
```golang
package main

import (
    "github.com/TarsCloud/TarsGo/tars"
    "helloworld" 
)

type GreeterImp  struct {
}

func (imp *GreeterImp) SayHello(input helloworld.HelloRequest)(output helloworld.HelloReply, err error) {
    output.Message = "hello" +  input.GetName() 
    return output, nil 
}

func main() {
    // Init servant
    imp := new(GreeterImp)                                    //New Imp
    app := new(helloworld.Greeter)                            //New init the A JCE
    cfg := tars.GetServerConfig()                              //Get Config File Object
    app.AddServant(imp, cfg.App+"."+cfg.Server+".GreeterTestObj") //Register Servant
    tars.Run()
}
```

- client
```golang
package main

import (
    "fmt"
    "github.com/TarsCloud/TarsGo/tars"
    "helloworld"
)

func main() {
    comm := tars.NewCommunicator()
    obj := fmt.Sprintf("StressTest.HelloPbServer.GreeterTestObj@tcp -h 127.0.0.1  -p 10014  -t 60000")
    app := new(helloworld.Greeter)
    comm.StringToProxy(obj, app)
    input := helloworld.HelloRequest{Name: "sandyskies"}
    output, err := app.SayHello(input)
    if err != nil {
        fmt.Println("err: ", err)
    }   
    fmt.Println("result is:", output.Message)
}
```

- config.conf
```xml
<tars>
    <application>
        <server>
            app=StressTest
            server=HelloPbServer
            local=tcp -h 127.0.0.1 -p 10014 -t 30000
            logpath=/tmp
            <StressTest.HelloPbServer.GreeterTestObjAdapter>
                allow
                endpoint=tcp -h 127.0.0.1 -p 10014 -t 60000
                handlegroup=StressTest.HelloPbServer.GreeterTestObjAdapter
                maxconns=200000
                protocol=tars
                queuecap=10000
                queuetimeout=60000
                servant=StressTest.HelloPbServer.GreeterTestObj
                shmcap=0
                shmkey=0
                threads=1
            </StressTest.HelloPbServer.GreeterTestObjAdapter>
        </server>
    </application>
</tars>
```