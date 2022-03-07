package main

import (
	"context"
	"fmt"
	"github.com/TarsCloud/TarsGo/tars"
	"trace/backend/tars-protocol/Trace"
)

// FrontendImp servant implementation
type FrontendImp struct {
	backend *Trace.Backend
}

// Init servant init
func (imp *FrontendImp) Init() error {
	//initialize servant here:
	comm := tars.NewCommunicator()
	obj := fmt.Sprintf("Trace.TarsTraceBackServer.BackendObj@tcp -h 127.0.0.1 -p 20015 -t 60000")
	app := new(Trace.Backend)
	comm.StringToProxy(obj, app)
	imp.backend = app
	return nil
}

// Destroy servant destory
func (imp *FrontendImp) Destroy() {
	//destroy servant here:
	//...
}

func (imp *FrontendImp) Add(ctx context.Context, a int32, b int32, c *int32) (int32, error) {
	//Doing something in your function
	return imp.backend.AddWithContext(ctx, a, b, c)
}
func (imp *FrontendImp) Sub(ctx context.Context, a int32, b int32, c *int32) (int32, error) {
	//Doing something in your function
	return imp.backend.SubWithContext(ctx, a, b, c)
}
