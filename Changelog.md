# Changelog

## 1.2.0 (2022/03/07)
- Support tars-trace @lbbniu #326
- support rpc keepalive @BeyondWUXF #3

## 1.1.9 (2022/01/24)
- 避免tarsgo每次创建项目都安装tars2go协议生成程序 by @lbbniu in #322
- 配置文件client objQueueMax不生效 by @BeyondWUXF in #323

## 1.1.8(2022/01/13)
- TarsGo客户端独立运行时支持stat和property自动上报 by @lbbniu in #314
- update tars version in setting.go by @lbbniu in #321

## 1.1.7 (2022/01/12)
- fix: JSON number loss of precision by @chenhengqi in #270
- Feature remotelog by @lanhy in #269
- fix: close fd when close connection by @chenhengqi in #274
- Support dispatch reporter by @defool in #282
- Set GOOS in make upload by @defool in #281
- Support setting the stater by @defool in #287
- Set default for json req value by @defool in #288
- 修正map没有初始化而直接使用时报assignment to entry in nil map错误的问题 by @tzwsoho in #286
- fix index by @BaBahsq in #293
- fix handletimeout ineffective by @Isites in #295
- Improve performance of codec by @defool in #301
- support call by set by @wgfxcu in #303
- correcting error output 'endoint' -> 'endpoint' by @Benjamin142857 in #308
- tars2go支持include参数 和 单文件定义多个模块 by @lbbniu in #296
- 优化单个tars文件中定义多个module处理 by @lbbniu in #316
- go 1.16+ 版本 创建项目使用脚手架tarsgo命令 by @lbbniu in #311
- 支持ssl拆分PR：对标c++支持https by @lbbniu in #319
- 一致性hash优化 by @lbbniu in #320

## 1.1.6 (2021/02/09)
- support rpc error code
- Support registering multiple filter for TarsServer and TarsClient.
- tars2go support: include support relative path
- add DoClose callback
- Fix panic when invoke timeout
- fixed response occasionally timeout
- fix missing formatting directives
- Do not show error when using tars as client
- support parent directory of outdir in tars2go
- fix timeout for recv
- mem use optimization
- replace expired page links with new links
- fix readme and setting.go comment typo
- fix report max numInvoke bug
- fix avoid create request id == 0
- fix md format
## 1.1.5 (2020/07/19)
- improve some features
- export `newServantProxy` function as public

## 1.1.4 (2020/06/15)
### ChangLogs
- merge from feature/tars-gateway branch (add tup && json protocol support)


## 1.1.3 (2020/06/02)
### ChangLogs
- fixed RequestPacket not set timeout bug
- fix README

## 1.1.2 (2020/04/23)
### ChangLogs
- merge from tars internal version

## 1.1.1 (2020/04/06)
### ChangLogs
- zipkin plugin version update (#172)
- fix reader missing on empty parameter list
- fixed the compile error package name upperFirstLatter (#179)
- fix(remotelogger): fix the issue that sync log to remote cause CPU usage too high.
- util/rogger: 1.showing func name without pkg name when logLevel is debug; 2.supported colored logLevel string when use console writer;
- style: goimports .
- style: make each line shorter
- style: fix typo
- style: gofmt .
- Fix empty-stringed-error when object gets Tars eroror code without messages.
- export method to get config and notify of ServerConfig (#150)
- Add Go-module (#135)
- statf: fix memory leak by removing mStatCount
- terminate tars decoding when missing required field
- ignore case when reading enableset config (#134)
- add configurable package length (#127)
- avoid compiling test code into application executable (#124)
- call flag.Parse if necessary, add udp client IP/port to context (#123)
- flag.Parse() must not be called during init. Instead, register flags during init, and call flag.Parse() in main(). (#113)
- fix nil config (#116)
- fix GetConfig fail on tars public cloud env (#104)
- fix create_tars_server.sh error @ubuntu (#98)
-  support grace restart (#95)
- modify logger for enable prefix (#97)
- tars2go support enum type
- get conf from template (#92)
- Fixed tars.reportNotifyInfo is not available;  and pull again (#86)
- fix deadlock
- fix ineffassign
- fix property nil panic
- add remote logger report interval
- Fix admin servant don't report notify msgs. (#80)
- Fix endpoint manager find activeEps error. (#79)
- fix golint for appprotocol.go
- Fixed a bug that make cleanall does not remove *.tgz files (#81)
- Fix endpoint manager find activeEps error. (#79)
- Fix adminservant don't report notify msgs. (#80)
- fix ineffectual assignment
- Modified # for config value that inside a line, only regard # as comment at the beginning. (#78)
- makefile support multi GOPATH (#76)
- Modified taf to tars in examples/EchoTestServer and added a client for it. (#75)
- go lint examples
- change error report
- fix gpool nil panic
- gofmt -s -w all

## 1.1.0 (2018/11/13)
### Feature
- Add contex support , put tarscurrent in context,for getting client ip ,port and so on.
- Add optional parameter for put context in request pacakge
- Add filter for writing plugin of tars service
- Add zipkin opentracing plugin
- Add support for protocol buffers


### Fix and enhancement.

- Change request package sbuffer field from vector<unsigned byte> to vector<byte>
- Fix stat report bug
- Getting Loglevel for remote configration
- Fix deadlock of getting routing infomation in extreme situation
- Improve goroutine pool 
- Fix occasionally panic problem because of the starting sequence of goroutines
- Golint most of the codes
