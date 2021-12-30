# 使用文档
- 安装 tarsgo
```bash
# < go 1.17 
go get -u github.com/TarsCloud/TarsGo/tars/tools/tarsgo
# >= go 1.17
go install github.com/TarsCloud/TarsGo/tars/tools/tarsgo@latest
```

- 帮助
```bash
$ tarsgo
tarsgo: An elegant toolkit for Go microservices.

Usage:
  tarsgo [command]

Available Commands:
  cmake       Create a service cmake template
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  make        Create a server make template

Flags:
  -h, --help      help for tarsgo
  -v, --version   version for tarsgo

Use "tarsgo [command] --help" for more information about a command.

$ tarsgo make help
Create a server make project using the repository template. Example:
tarsgo make TeleSafe PhonenumSogouServer SogouInfo github.com/TeleSafe/PhonenumSogouServer

Usage:
  tarsgo make App Server Servant GoModuleName [flags]

Flags:
  -b, --branch string     repo branch
  -h, --help              help for make
  -r, --repo-url string   layout repo (default "https://github.com/TarsCloud/TarsGo.git")
  -t, --timeout string    time out (default "60s")
  
$ tarsgo cmake help
Create a server make project using the repository template. Example:
tarsgo make TeleSafe PhonenumSogouServer SogouInfo github.com/TeleSafe/PhonenumSogouServer

Usage:
  tarsgo make App Server Servant GoModuleName [flags]

Flags:
  -b, --branch string     repo branch
  -h, --help              help for make
  -r, --repo-url string   layout repo (default "https://github.com/TarsCloud/TarsGo.git")
  -t, --timeout string    time out (default "60s")
(base) ➜  src tarsgo cmake help
Create a service cmake project using the repository template. Example:
tarsgo cmake TeleSafe PhonenumSogouServer SogouInfo github.com/TeleSafe/PhonenumSogouServer

Usage:
  tarsgo cmake App Server Servant GoModuleName [flags]

Flags:
  -b, --branch string     repo branch
  -h, --help              help for cmake
  -r, --repo-url string   layout repo (default "https://github.com/TarsCloud/TarsGo.git")
  -t, --timeout string    time out (default "60s")
```

- 创建make管理项目
```bash
$ tarsgo make TeleSafe PhonenumSogouServer SogouInfo github.com/TeleSafe/PhonenumSogouServer
go get -u github.com/TarsCloud/TarsGo/tars/tools/tars2go
🚀 Creating server TeleSafe.PhonenumSogouServer, layout repo is https://github.com/TarsCloud/TarsGo.git, please wait a moment.

正克隆到 '/Users/xxx/.tarsgo/repo/github.com/TarsCloud/TarsGo@main'...

go: creating new go.mod: module github.com/TeleSafe/PhonenumSogouServer
go: to add module requirements and sums:
	go mod tidy

CREATED PhonenumSogouServer/SogouInfo.tars (173 bytes)
CREATED PhonenumSogouServer/SogouInfo_imp.go (628 bytes)
CREATED PhonenumSogouServer/client/client.go (426 bytes)
CREATED PhonenumSogouServer/config.conf (1036 bytes)
CREATED PhonenumSogouServer/debugtool/dumpstack.go (439 bytes)
CREATED PhonenumSogouServer/go.mod (56 bytes)
CREATED PhonenumSogouServer/main.go (543 bytes)
CREATED PhonenumSogouServer/makefile (233 bytes)
CREATED PhonenumSogouServer/scripts/makefile.tars.gomod (4165 bytes)
CREATED PhonenumSogouServer/start.sh (60 bytes)

>>> Great！Done! You can jump in PhonenumSogouServer
>>> Tips: After editing the Tars file, execute the following cmd to automatically generate golang files.
>>>       /Users/xxx/go/bin/tars2go *.tars
$ cd PhonenumSogouServer
$ ./start.sh
🤝 Thanks for using TarsGo
📚 Tutorial: https://tarscloud.github.io/TarsDocs/
```

- 创建cmake管理项目
```bash
$ tarsgo cmake TeleSafe PhonenumSogouServer SogouInfo github.com/TeleSafe/PhonenumSogouServer
go get -u github.com/TarsCloud/TarsGo/tars/tools/tars2go
🚀 Creating server TeleSafe.PhonenumSogouServer, layout repo is https://github.com/TarsCloud/TarsGo.git, please wait a moment.

已经是最新的。

go: creating new go.mod: module github.com/TeleSafe/PhonenumSogouServer
go: to add module requirements and sums:
	go mod tidy

CREATED PhonenumSogouServer/CMakeLists.txt (416 bytes)
CREATED PhonenumSogouServer/SogouInfo.tars (173 bytes)
CREATED PhonenumSogouServer/SogouInfo_imp.go (628 bytes)
CREATED PhonenumSogouServer/client/CMakeLists.txt (164 bytes)
CREATED PhonenumSogouServer/client/client.go (426 bytes)
CREATED PhonenumSogouServer/cmake/CMakeDetermineGoCompiler.cmake (1615 bytes)
CREATED PhonenumSogouServer/cmake/CMakeGoCompiler.cmake.in (273 bytes)
CREATED PhonenumSogouServer/cmake/CMakeGoInformation.cmake (230 bytes)
CREATED PhonenumSogouServer/cmake/CMakeTestGoCompiler.cmake (49 bytes)
CREATED PhonenumSogouServer/cmake/golang.cmake (1993 bytes)
CREATED PhonenumSogouServer/cmake/tars-tools.cmake (9754 bytes)
CREATED PhonenumSogouServer/config.conf (1036 bytes)
CREATED PhonenumSogouServer/debugtool/dumpstack.go (439 bytes)
CREATED PhonenumSogouServer/go.mod (56 bytes)
CREATED PhonenumSogouServer/main.go (543 bytes)
CREATED PhonenumSogouServer/start.sh (66 bytes)

>>> Great！Done! You can jump in PhonenumSogouServer
>>> Tips: After editing the Tars file, execute the following cmd to automatically generate golang files.
>>>       /Users/xxx/go/bin/tars2go *.tars
$ cd PhonenumSogouServer
$ ./start.sh
🤝 Thanks for using TarsGo
📚 Tutorial: https://tarscloud.github.io/TarsDocs/
```