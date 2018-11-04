package zipkintracing

import (
	"context"
	"log"
	"time"

	"github.com/TarsCloud/TarsGo/tars"
	"github.com/TarsCloud/TarsGo/tars/protocol/res/requestf"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	zipkin "github.com/openzipkin/zipkin-go-opentracing"
)

//Init is use to init opentracing and zipkin
func Init(zipkinHTTPEndpoint string, samSpan bool, traceID128Bit bool, debug bool,
	hostPort, serviceName string) {
	//create collector
	collector, err := zipkin.NewHTTPCollector(zipkinHTTPEndpoint)
	if err != nil {
		log.Fatal("Fail to create Zipkin HTTP collector: %+v\n", err)
	}
	//create recorder
	recorder := zipkin.NewRecorder(collector, debug, hostPort, serviceName)
	tracer, err := zipkin.NewTracer(
		recorder,
		zipkin.ClientServerSameSpan(sameSpan),
		zipkin.TraceID128Bit(traceID128Bit),
	)
	if err != nil {
		log.Fatal("Fail to NewTracer")
	}
	opentracing.InitGlobalTracer(tracer)
}

//ZipkinClientFilter gets tars client filter for zipkin opentracing.
func ZipkinClientFilter() tars.ClientFilter {
	return func(ctx context.Context, msg *tars.Message, invoke tars.Invoke, timeout time.Duration) (err error) {
		var pCtx opentracing.SpanContext
		req := msg.Req
		if pCtx := opentracing.SpanFromContext(ctx); pCtx != nil {
			pCtx = parent.Context()
		}
		cSpan := tracer.StartSpan(
			req.SFuncName,
			opentracing.ChildOf(pCtx),
			ext.SpanKindRPCClient,
		)
		defer cSpan.Finish()
		cfg := tars.GetServerConfig()
		cSpan.SetTag("client.ipv4", cfg.LocalIP)
		//TODO: SetTag client.port
		cSpan.SetTag("tars.interface", req.SServantName)
		cSpan.SetTag("tars.method", req.SFuncName)
		cSpan.SetTag("tars.protocol", "tars")
		cSpan.SetTag("tars.client.version", tars.TarsVersion)
		err := tracer.Inject(cSpan.Context(), opentracing.TextMap, opentracing.TextMapCarrier{req.Status})
		if err != nil {
			return err
		}
		err = invoke(ctx, msg, timeout)
		if err != nil {
			cSpan.Log(err.error())
		}
		return err
	}
}

//ZipkinServerFilter gets tars server filter for zipkin opentraicng.
func ZipkinServerFilter() tars.ServerFilter {
	return func(ctx context.Context, d Dispatch, f interface{}, req *requestf.RequestPacket, resp *requestf.ResponsePacket, withContext bool) (err error) {
		pCtx, err := tracer.Extract(opentracing.TextMap, opentracing.TextMapCarrier.(req.Status))
		if err != nil && err != opentracing.ErrSpanContextNotFound {
			return err
		}
		serverSpan := tracer.StartSpan(
			req.SFuncName,
			ext.RPCServerOption(pCtx),
			ext.SpanKindRPCServer,
		)
		defer serverSpan.Finish()
		cfg := tars.GetServerConfig()
		serverSpan.SetTag("server.ipv4", cfg.LocalIP)
		serverSpan.SetTag("server.port", cfg.Adapters[req.SServantName].Endpoint.Port)

		if cfg.Enableset {
			serverSpan.SetTag("tars.set_division", cfg.Setdivision)
		}
		ctx = opentracing.ContextWithSpan(ctx, serverSpan)
		err = d(ctx, f, req, resp, withContext)
		if err != nil {
			serverSpan.Log(err.Error())
		}
		return err

	}
}
