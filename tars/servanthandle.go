package tars

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"crypto/tls"	

	"github.com/TarsCloud/TarsGo/tars/transport"
)

//AddServant add dispatch and interface for object.
func AddServant(v dispatch, f interface{}, obj string) {
	addServantCommon(v, f, obj, false)
}

//AddServantWithContext add dispatch and interface for object, which have ctx,context
func AddServantWithContext(v dispatch, f interface{}, obj string) {
	addServantCommon(v, f, obj, true)
}

func addServantCommon(v dispatch, f interface{}, obj string, withContext bool) {
	objRunList = append(objRunList, obj)
	cfg, ok := tarsConfig[obj]
	if !ok {
		ReportNotifyInfo(NOTIFY_ERROR, "servant obj name not found:"+obj)
		TLOG.Debug("servant obj name not found:", obj)
		panic(ok)
	}
	if v, ok := f.(destroyableImp); ok {
		TLOG.Debugf("add destroyable obj %s", obj)
		destroyableObjs = append(destroyableObjs, v)
	}
	TLOG.Debug("add:", cfg)

	jp := NewTarsProtocol(v, f, withContext)
	s := transport.NewTarsServer(jp, cfg)
	goSvrs[obj] = s
}

// AddHttpServant add http servant handler with default exceptionStatusChecker for obj.
func AddHttpServant(mux *TarsHttpMux, obj string) {
	AddHttpServantWithExceptionStatusChecker(mux, obj, DefaultExceptionStatusChecker)
}

// AddHttpServantWithExceptionStatusChecker add http servant handler with exceptionStatusChecker for obj.
func AddHttpServantWithExceptionStatusChecker(mux *TarsHttpMux, obj string, exceptionStatusChecker func(int) bool) {
	cfg, ok := tarsConfig[obj]
	if !ok {
		ReportNotifyInfo(NOTIFY_ERROR, "servant obj name not found:"+obj)
		TLOG.Debug("servant obj name not found:", obj)
		panic(ok)
	}
	TLOG.Debug("add http server:", cfg)
	objRunList = append(objRunList, obj)
	appConf := GetServerConfig()
	addrInfo := strings.SplitN(cfg.Address, ":", 2)
	var port int
	if len(addrInfo) == 2 {
		port, _ = strconv.Atoi(addrInfo[1])
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
	s := &http.Server{Addr: cfg.Address, Handler: mux}
	httpSvrs[obj] = s
}

// AddServantWithProtocol adds a servant with protocol and obj
func AddServantWithProtocol(proto transport.ServerProtocol, obj string) {
	objRunList = append(objRunList, obj)
	cfg, ok := tarsConfig[obj]
	if !ok {
		ReportNotifyInfo(NOTIFY_ERROR, "servant obj name not found:"+obj)
		TLOG.Debug("servant obj name not found ", obj)
		panic(ok)
	}
	TLOG.Debug("add:", cfg)
	s := transport.NewTarsServer(proto, cfg)
	goSvrs[obj] = s
}

// AddHttpsServant add https servant handler with default exceptionStatusChecker for obj.
func AddHttpsServant(mux *TarsHttpMux, obj, certFile, keyFile string) {
	AddHttpsServantWithExceptionStatusChecker(mux, obj, certFile, keyFile, DefaultExceptionStatusChecker)
}

// AddHttpsServantWithExceptionStatusChecker add https servant handler with exceptionStatusChecker for obj.
func AddHttpsServantWithExceptionStatusChecker(mux *TarsHttpMux, obj, certFile, keyFile string, exceptionStatusChecker func(int) bool) {
	cfg, ok := tarsConfig[obj]
	if !ok {
		ReportNotifyInfo(NOTIFY_ERROR, "servant obj name not found:"+obj)
		TLOG.Debug("servant obj name not found:", obj)
		panic(ok)
	}
	TLOG.Debug("add http server:", cfg)
	objRunList = append(objRunList, obj)
	appConf := GetServerConfig()
	addrInfo := strings.SplitN(cfg.Address, ":", 2)
	var port int
	if len(addrInfo) == 2 {
		port, _ = strconv.Atoi(addrInfo[1])
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
	s := &http.Server{Addr: cfg.Address, Handler: mux}
	s.TLSConfig = &tls.Config{Certificates:make([]tls.Certificate, 1)}
	var err error
	s.TLSConfig.Certificates[0], err = tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		ReportNotifyInfo(NOTIFY_ERROR, "servant can't load X509 Key Pair from file:"+certFile+","+keyFile)
		TLOG.Debug("servant can't load X509 Key Pair from file:"+certFile+","+keyFile)
		panic(err)
	}
	httpSvrs[obj] = s
}
