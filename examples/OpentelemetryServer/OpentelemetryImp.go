package main

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/TarsCloud/TarsGo/tars"
	"github.com/TarsCloud/TarsGo/tars/util/current"
)

var name = "OpentelemetryImp"

type OpentelemetryImp struct {
}

func (imp *OpentelemetryImp) Add(ctx context.Context, a int32, b int32, c *int32) (int32, error) {
	span := trace.SpanContextFromContext(ctx)
	logger := tars.GetLogger("context")
	data, err := span.MarshalJSON()
	if err != nil {
		logger.Errorf("MarshalJSON err：%v", err)
	}
	logger.Info("span", string(data))
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
	if !ok {
		logger.Error("error setting respose context")
	}
	imp.Sub(ctx, a, b, c)
	//Doing something in your function
	//...
	*c = a * b
	return 0, nil
}

func (imp *OpentelemetryImp) Sub(ctx context.Context, a int32, b int32, c *int32) (int32, error) {
	_, span := otel.Tracer(name).Start(ctx, "Sub")
	defer span.End()
	span.SetAttributes(attribute.Int64("request.a", int64(a)))
	span.SetAttributes(attribute.Int64("request.b", int64(b)))
	//...
	return 0, nil
}
