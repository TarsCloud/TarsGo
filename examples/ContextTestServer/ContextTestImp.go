package main

import (
	"context"

	"github.com/TarsCloud/TarsGo/tars"

	"github.com/TarsCloud/TarsGo/tars/util/current"
)

type ContextTestImp struct {
}

func (imp *ContextTestImp) Add(ctx context.Context, a int32, b int32, c *int32) (int32, error) {
	logger := tars.GetLogger("context")
	ip, ok := current.GetClientIPFromContext(ctx)
	if !ok {
		logger.Error("Error getting ip from context")
	}
	logger.Infof("Get Client Ip : %s from context", ip)
	reqContext, ok := current.GetRequestContext(ctx)
	if !ok {
		logger.Error("Error getting reqcontext from context")
	}
	logger.Infof("Get context from context: %v", reqContext)
	k := make(map[string]string)
	k["resp"] = "respform context"
	ok = current.SetResponseContext(ctx, k)
	//Doing something in your function
	//...
	*c = a * b
	return 0, nil
}
func (imp *ContextTestImp) Sub(ctx context.Context, a int32, b int32, c *int32) (int32, error) {
	//Doing something in your function
	//...
	return 0, nil
}
