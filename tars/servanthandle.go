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
		TLOG.Debug("servant obj name not found ", obj)
		return
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

//AddHttpServant add http servant handler for obj.
func AddHttpServant(mux *TarsHttpMux, obj string) {
	cfg, ok := tarsConfig[obj]
	if !ok {
		TLOG.Debug("servant obj name not found ", obj)
		return
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
		Container: appConf.Container,
		AppName:   fmt.Sprintf("%s.%s", appConf.App, appConf.Server),
		Version:   appConf.Version,
		IP:        addrInfo[0],
		Port:      int32(port),
		SetId:     appConf.Setdivision,
	}
	mux.SetConfig(httpConf)
	s := &http.Server{Addr: cfg.Address, Handler: mux}
	httpSvrs[obj] = s
}

//AddHttpsServant add https servant handler for obj.
func AddHttpsServant(mux *TarsHttpMux, obj, certFile, keyFile string) {
	cfg, ok := tarsConfig[obj]
	if !ok {
		TLOG.Debug("servant obj name not found ", obj)
		return
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
		Container: appConf.Container,
		AppName:   fmt.Sprintf("%s.%s", appConf.App, appConf.Server),
		Version:   appConf.Version,
		IP:        addrInfo[0],
		Port:      int32(port),
		SetId:     appConf.Setdivision,
	}
	mux.SetConfig(httpConf)
	s := &http.Server{Addr: cfg.Address, Handler: mux}
	s.TLSConfig = &tls.Config{Certificates:make([]tls.Certificate, 1)}
	var err error
	s.TLSConfig.Certificates[0], err = tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		TLOG.Debug("servant can't load X509 Key Pair from file:"+certFile+","+keyFile)
		return
	}	
	httpSvrs[obj] = s
}
