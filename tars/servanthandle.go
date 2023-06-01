package tars

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/TarsCloud/TarsGo/tars/transport"
)

// AddServant add dispatch and interface for object.
func AddServant(v dispatch, f interface{}, obj string) {
	defaultApp.AddServant(v, f, obj)
}

// AddServantWithContext add dispatch and interface for object, which have ctx,context
func AddServantWithContext(v dispatch, f interface{}, obj string) {
	defaultApp.AddServantWithContext(v, f, obj)
}

// AddHttpServant add http servant handler with default exceptionStatusChecker for obj.
func AddHttpServant(mux HttpHandler, obj string) {
	AddHttpServantWithExceptionStatusChecker(mux, obj, DefaultExceptionStatusChecker)
}

// AddHttpServantWithExceptionStatusChecker add http servant handler with exceptionStatusChecker for obj.
func AddHttpServantWithExceptionStatusChecker(mux HttpHandler, obj string, exceptionStatusChecker func(int) bool) {
	defaultApp.AddHttpServantWithExceptionStatusChecker(mux, obj, exceptionStatusChecker)
}

// AddServantWithProtocol adds a servant with protocol and obj
func AddServantWithProtocol(proto transport.ServerProtocol, obj string) {
	defaultApp.AddServantWithProtocol(proto, obj)
}

// AddServant add dispatch and interface for object.
func (a *application) AddServant(v dispatch, f interface{}, obj string) {
	a.addServantCommon(v, f, obj, false)
}

// AddServantWithContext add dispatch and interface for object, which have ctx,context
func (a *application) AddServantWithContext(v dispatch, f interface{}, obj string) {
	a.addServantCommon(v, f, obj, true)
}

func (a *application) addServantCommon(v dispatch, f interface{}, obj string, withContext bool) {
	a.objRunList = append(a.objRunList, obj)
	cfg, ok := a.tarsConfig[obj]
	if !ok {
		msg := fmt.Sprintf("tars servant obj name not found: %s", obj)
		ReportNotifyInfo(NotifyError, msg)
		TLOG.Debug(msg)
		panic(errors.New(msg))
	}
	if v, ok := f.(destroyableImp); ok {
		TLOG.Debugf("add destroyable obj %s", obj)
		a.destroyableObjs = append(a.destroyableObjs, v)
	}
	TLOG.Debugf("add tars protocol server: %+v", cfg)

	jp := NewTarsProtocol(v, f, withContext)
	jp.app = a
	s := transport.NewTarsServer(jp, cfg)
	a.goSvrs[obj] = s
}

// AddHttpServant add http servant handler with default exceptionStatusChecker for obj.
func (a *application) AddHttpServant(mux HttpHandler, obj string) {
	a.AddHttpServantWithExceptionStatusChecker(mux, obj, DefaultExceptionStatusChecker)
}

// AddHttpServantWithExceptionStatusChecker add http servant handler with exceptionStatusChecker for obj.
func (a *application) AddHttpServantWithExceptionStatusChecker(mux HttpHandler, obj string, exceptionStatusChecker func(int) bool) {
	cfg, ok := a.tarsConfig[obj]
	if !ok {
		msg := fmt.Sprintf("http servant obj name not found: %s", obj)
		ReportNotifyInfo(NotifyError, msg)
		TLOG.Debug(msg)
		panic(errors.New(msg))
	}
	TLOG.Debugf("add http protocol server: %+v", cfg)
	a.objRunList = append(a.objRunList, obj)
	addrInfo := strings.SplitN(cfg.Address, ":", 2)
	var port int64
	if len(addrInfo) == 2 {
		var err error
		port, err = strconv.ParseInt(addrInfo[1], 10, 32)
		if err != nil {
			panic(fmt.Errorf("http server listen port: %s parse err: %v", addrInfo[1], err))
		}
	}
	svrCfg := a.ServerConfig()
	httpConf := &TarsHttpConf{
		Container:              svrCfg.Container,
		AppName:                fmt.Sprintf("%s.%s", svrCfg.App, svrCfg.Server),
		Version:                svrCfg.Version,
		IP:                     addrInfo[0],
		Port:                   int32(port),
		SetId:                  svrCfg.Setdivision,
		ExceptionStatusChecker: exceptionStatusChecker,
	}
	mux.SetConfig(httpConf)
	s := &http.Server{Addr: cfg.Address, Handler: mux, TLSConfig: cfg.TlsConfig}
	a.httpSvrs[obj] = s
}

// AddServantWithProtocol adds a servant with protocol and obj
func (a *application) AddServantWithProtocol(proto transport.ServerProtocol, obj string) {
	a.objRunList = append(a.objRunList, obj)
	cfg, ok := a.tarsConfig[obj]
	if !ok {
		msg := fmt.Sprintf("custom protocol servant obj name not found: %s", obj)
		ReportNotifyInfo(NotifyError, msg)
		TLOG.Debug(msg)
		panic(errors.New(msg))
	}
	TLOG.Debugf("add custom protocol server: %+v", cfg)
	s := transport.NewTarsServer(proto, cfg)
	a.goSvrs[obj] = s
}
