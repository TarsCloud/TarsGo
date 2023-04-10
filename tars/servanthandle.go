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
func AddServant(v dispatch, f any, obj string) {
	addServantCommon(v, f, obj, false)
}

// AddServantWithContext add dispatch and interface for object, which have ctx,context
func AddServantWithContext(v dispatch, f any, obj string) {
	addServantCommon(v, f, obj, true)
}

func addServantCommon(v dispatch, f any, obj string, withContext bool) {
	objRunList = append(objRunList, obj)
	cfg, ok := tarsConfig[obj]
	if !ok {
		msg := fmt.Sprintf("tars servant obj name not found: %s", obj)
		ReportNotifyInfo(NotifyError, msg)
		TLOG.Debug(msg)
		panic(errors.New(msg))
	}
	if v, ok := f.(destroyableImp); ok {
		TLOG.Debugf("add destroyable obj %s", obj)
		destroyableObjs = append(destroyableObjs, v)
	}
	TLOG.Debug("add tars protocol server: %+v", cfg)

	jp := NewTarsProtocol(v, f, withContext)
	s := transport.NewTarsServer(jp, cfg)
	goSvrs[obj] = s
}

// AddHttpServant add http servant handler with default exceptionStatusChecker for obj.
func AddHttpServant(mux HttpHandler, obj string) {
	AddHttpServantWithExceptionStatusChecker(mux, obj, DefaultExceptionStatusChecker)
}

// AddHttpServantWithExceptionStatusChecker add http servant handler with exceptionStatusChecker for obj.
func AddHttpServantWithExceptionStatusChecker(mux HttpHandler, obj string, exceptionStatusChecker func(int) bool) {
	cfg, ok := tarsConfig[obj]
	if !ok {
		msg := fmt.Sprintf("http servant obj name not found: %s", obj)
		ReportNotifyInfo(NotifyError, msg)
		TLOG.Debug(msg)
		panic(errors.New(msg))
	}
	TLOG.Debugf("add http protocol server: %+v", cfg)
	objRunList = append(objRunList, obj)
	appConf := GetServerConfig()
	addrInfo := strings.SplitN(cfg.Address, ":", 2)
	var port int64
	if len(addrInfo) == 2 {
		var err error
		port, err = strconv.ParseInt(addrInfo[1], 10, 32)
		if err != nil {
			panic(fmt.Errorf("http server listen port: %s parse err: %v", addrInfo[1], err))
		}
	}
	httpConf := &TarsHttpConf{
		Container:              appConf.Container,
		AppName:                fmt.Sprintf("%s.%s", appConf.App, appConf.Server),
		Version:                appConf.Version,
		IP:                     addrInfo[0],
		Port:                   int32(port),
		SetId:                  appConf.Setdivision,
		ExceptionStatusChecker: exceptionStatusChecker,
	}
	mux.SetConfig(httpConf)
	s := &http.Server{Addr: cfg.Address, Handler: mux, TLSConfig: cfg.TlsConfig}
	httpSvrs[obj] = s
}

// AddServantWithProtocol adds a servant with protocol and obj
func AddServantWithProtocol(proto transport.ServerProtocol, obj string) {
	objRunList = append(objRunList, obj)
	cfg, ok := tarsConfig[obj]
	if !ok {
		msg := fmt.Sprintf("custom protocol servant obj name not found: %s", obj)
		ReportNotifyInfo(NotifyError, msg)
		TLOG.Debug(msg)
		panic(errors.New(msg))
	}
	TLOG.Debugf("add custom protocol server: %+v", cfg)
	s := transport.NewTarsServer(proto, cfg)
	goSvrs[obj] = s
}
