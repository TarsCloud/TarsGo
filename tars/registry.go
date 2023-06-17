package tars

import (
	"context"

	"github.com/TarsCloud/TarsGo/tars/registry"
	"github.com/TarsCloud/TarsGo/tars/util/endpoint"
)

func (a *application) registryAdapters(ctx context.Context) {
	if a.opt.registrar == nil {
		return
	}
	svrCfg := GetServerConfig()
	for _, adapter := range svrCfg.Adapters {
		servant := &registry.ServantInstance{
			TarsVersion: Version,
			App:         svrCfg.App,
			Server:      svrCfg.Server,
			Servant:     adapter.Obj,
			EnableSet:   svrCfg.Enableset,
			SetDivision: svrCfg.Setdivision,
			Protocol:    adapter.Protocol,
			Endpoint:    endpoint.Endpoint2tars(adapter.Endpoint),
		}
		if err := a.opt.registrar.Registry(ctx, servant); err != nil {
			TLOG.Errorf("registry %+v error: %+v", servant, err)
		}
	}
}

func (a *application) deregisterAdapters(ctx context.Context) {
	if a.opt.registrar == nil {
		return
	}
	svrCfg := GetServerConfig()
	for _, adapter := range svrCfg.Adapters {
		servant := &registry.ServantInstance{
			TarsVersion: Version,
			App:         svrCfg.App,
			Server:      svrCfg.Server,
			EnableSet:   svrCfg.Enableset,
			SetDivision: svrCfg.Setdivision,
			Servant:     adapter.Obj,
			Protocol:    adapter.Protocol,
			Endpoint:    endpoint.Endpoint2tars(adapter.Endpoint),
		}
		if err := a.opt.registrar.Deregister(ctx, servant); err != nil {
			TLOG.Errorf("deregister: %+v error: %+v", servant, err)
		}
	}
}
