package base

import (
	"bytes"
	"os"
	"path"
	"strings"
	"text/template"

	"github.com/TarsCloud/TarsGo/tars/tools/tarsgo/internal/bindata"
	"github.com/TarsCloud/TarsGo/tars/tools/tarsgo/internal/consts"
)

const (
	cmakeStartTemplateFile = "start.sh"
	cmakeStartTemplate     = `#!/bin/bash
set -ex
mkdir -p build
cd build
cmake ..
make
cd -
./build/bin/{{.Server}} --config=config/config.conf
`
	cmakeListsTxtTemplateFile = "CMakeLists.txt"
	cmakeListsTxtTemplate     = `
execute_process(COMMAND go env GOPATH OUTPUT_VARIABLE GOPATH)
string(REGEX REPLACE "\n$" "" GOPATH "${GOPATH}")

include(${CMAKE_CURRENT_SOURCE_DIR}/cmake/tars-tools.cmake)

cmake_minimum_required(VERSION 2.8)

project({{.Server}} Go) # select GO compile

gen_server({{.App}} {{.Server}})

add_subdirectory(client)

# go env -w GO111MODULE=on
# go mod init
# mkdir build
# cd build
# cmake ..
# make`

	clientCMakeListsTxtTemplateFile = "client/CMakeLists.txt"
	clientCMakeListsTxtTemplate     = `
cmake_minimum_required(VERSION 2.8)

project(client Go) # select GO compile

gen_server({{.App}} client) 

# go mod init
# mkdir build
# cd build
# cmake ..
# make`

	makeStartTemplateFile = "start.sh"
	makeStartTemplate     = `#!/bin/bash
set -ex
make
./{{.Server}} --config=config/config.conf
`

	makefileTemplateFile = "Makefile"
	makefileTemplate     = `APP       := {{.App}}
TARGET    := {{.Server}}
MFLAGS    :=
DFLAGS    :=
CONFIG    := client
STRIP_FLAG:= N
J2GO_FLAG:= 

-include scripts/makefile.tars.gomod.mk
`

	mainGoTemplateFile = "main.go"
	mainGoTemplate     = `package main

import (
	"fmt"
	"os"

	"github.com/TarsCloud/TarsGo/tars"

	"{{.GoModuleName}}/tars-protocol/{{.App}}"
)

func main() {
	// Get server config
	cfg := tars.GetServerConfig()

	// New servant imp
	imp := new({{.Servant}}Imp)
	err := imp.Init()
	if err != nil {
		fmt.Printf("{{.Servant}}Imp init fail, err:(%s)\n", err)
		os.Exit(-1)
	}
	// New servant
	app := new({{.App}}.{{.Servant}})
	// Register Servant
	app.AddServantWithContext(imp, cfg.App+"."+cfg.Server+".{{.Servant}}Obj")

	// Run application
	tars.Run()
}
`
	clientGoTemplateFile = "client/client.go"
	clientGoTemplate     = `package main

import (
	"fmt"

	"github.com/TarsCloud/TarsGo/tars"

	"{{.GoModuleName}}/tars-protocol/{{.App}}"
)

func main() {
	comm := tars.NewCommunicator()
	obj := fmt.Sprintf("{{.App}}.{{.Server}}.{{.Servant}}Obj@tcp -h 127.0.0.1 -p 10015 -t 60000")
	app := new({{.App}}.{{.Servant}})
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
`
	configConfTemplateFile = "config/config.conf"
	configConfTemplate     = `<tars>
    <application>
        <server>
            app={{.App}}
            server={{.Server}}
            local=tcp -h 127.0.0.1 -p 10014 -t 30000
            logpath=/tmp
            <{{.App}}.{{.Server}}.{{.Servant}}ObjAdapter>
                allow
                endpoint=tcp -h 127.0.0.1 -p 10015 -t 60000
                handlegroup={{.App}}.{{.Server}}.{{.Servant}}ObjAdapter
                maxconns=200000
                protocol=tars
                queuecap=10000
                queuetimeout=60000
                servant={{.App}}.{{.Server}}.{{.Servant}}Obj
                shmcap=0
                shmkey=0
                threads=1
            </{{.App}}.{{.Server}}.{{.Servant}}ObjAdapter>
        </server>
    </application>
</tars>
`
	servantTarsTemplateFile = "Servant.tars"
	servantTarsTemplate     = `module {{.App}}
{
	interface {{.Servant}}
	{
	    int Add(int a,int b,out int c); // Some example function
	    int Sub(int a,int b,out int c); // Some example function
	};
};
`
	servantImpGoTemplateFile = "Servant_imp.go"
	servantImpGoTemplate     = `package main

import (
	"context"
)

// {{.Servant}}Imp servant implementation
type {{.Servant}}Imp struct {
}

// Init servant init
func (imp *{{.Servant}}Imp) Init() error {
	//initialize servant here:
	//...
	return nil
}

// Destroy servant destroy
func (imp *{{.Servant}}Imp) Destroy() {
	//destroy servant here:
	//...
}

func (imp *{{.Servant}}Imp) Add(ctx context.Context, a int32, b int32, c *int32) (int32, error) {
	//Doing something in your function
	//...
	return 0, nil
}
func (imp *{{.Servant}}Imp) Sub(ctx context.Context, a int32, b int32, c *int32) (int32, error) {
	//Doing something in your function
	//...
	return 0, nil
}
`

	dumpstackTemplateFile = "debugtool/dumpstack.go"
	dumpstackTemplate     = `package main

import (
	"fmt"

	"github.com/TarsCloud/TarsGo/tars"
	"github.com/TarsCloud/TarsGo/tars/protocol/res/adminf"
)

func main() {
	comm := tars.NewCommunicator()
	obj := "{{.App}}.{{.Server}}.{{.Servant}}Obj@tcp -h 127.0.0.1 -p 10014 -t 60000"
	app := new(adminf.AdminF)
	comm.StringToProxy(obj, app)
	ret, err := app.Notify("tars.dumpstack")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(ret)
}
`
)

var makeTemplates = map[string]string{
	makeStartTemplateFile:    makeStartTemplate,
	makefileTemplateFile:     makefileTemplate,
	configConfTemplateFile:   configConfTemplate,
	mainGoTemplateFile:       mainGoTemplate,
	clientGoTemplateFile:     clientGoTemplate,
	servantTarsTemplateFile:  servantTarsTemplate,
	servantImpGoTemplateFile: servantImpGoTemplate,
	dumpstackTemplateFile:    dumpstackTemplate,
}

var cmakeTemplates = map[string]string{
	cmakeStartTemplateFile:          cmakeStartTemplate,
	cmakeListsTxtTemplateFile:       cmakeListsTxtTemplate,
	clientCMakeListsTxtTemplateFile: clientCMakeListsTxtTemplate,
	configConfTemplateFile:          configConfTemplate,
	mainGoTemplateFile:              mainGoTemplate,
	clientGoTemplateFile:            clientGoTemplate,
	servantTarsTemplateFile:         servantTarsTemplate,
	servantImpGoTemplateFile:        servantImpGoTemplate,
	dumpstackTemplateFile:           dumpstackTemplate,
}

func DoGenProject(p *Project, to string, mgrType string) error {
	templates := makeTemplates
	if mgrType == consts.CMake {
		templates = cmakeTemplates
	}
	for filename, content := range templates {
		filename = path.Join(to, filename)
		if err := os.MkdirAll(path.Dir(filename), os.ModePerm); err != nil {
			return err
		}
		t := template.Must(template.New(filename).Parse(content))
		buffer := new(bytes.Buffer)
		err := t.Execute(buffer, p)
		if err != nil {
			return err
		}
		if strings.HasSuffix(filename, ".sh") {

			err = os.WriteFile(filename, buffer.Bytes(), os.ModePerm)
		} else {
			err = os.WriteFile(filename, buffer.Bytes(), 0666)
		}
		if err != nil {
			return err
		}
	}
	if mgrType == consts.CMake {
		return bindata.RestoreAssets(to, "cmake")
	}
	return bindata.RestoreAsset(to, "scripts/makefile.tars.gomod.mk")
}
