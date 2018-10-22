package tars

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/TarsCloud/TarsGo/tars/transport"
)

// AddServant adds a new servant
func AddServant(dispatcher Dispatcher, implementer Implementer, name string) error {
	if mux, ok := dispatcher.(*TarsHttpMux); ok {
		return addHttpServant(mux, name)
	} else if v, ok := dispatcher.(dispatch); ok {
		return addTUServant(v, implementer, name)
	} else {
		return fmt.Errorf("unsupported servant: %s", name)
	}
}

func addTUServant(v dispatch, f interface{}, obj string) error {
	objRunList = append(objRunList, obj)
	cfg, ok := tarsConfig[obj]
	if !ok {
		TLOG.Debug("servant obj name not found ", obj)
	}
	TLOG.Debug("add:", cfg)
	jp := NewTarsProtocol(v, f)
	s := transport.NewTarsServer(jp, cfg)
	goSvrs[obj] = s
	return nil
}

func addHttpServant(mux *TarsHttpMux, obj string) error {
	cfg, ok := tarsConfig[obj]
	if !ok {
		TLOG.Debug("servant obj name not found ", obj)
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
	return nil
}
