package zipkintracing

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/TarsCloud/TarsGo/tars"
	"github.com/TarsCloud/TarsGo/tars/protocol/res/requestf"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	oplog "github.com/opentracing/opentracing-go/log"
	zipkinot "github.com/openzipkin-contrib/zipkin-go-opentracing"
	"github.com/openzipkin/zipkin-go"
	zipkinhttp "github.com/openzipkin/zipkin-go/reporter/http"
)

var logger = tars.GetLogger("tracing")

//Init is use to init opentracing and zipkin
func Init(zipkinHTTPEndpoint string, sameSpan bool, traceID128Bit bool, debug bool,
	hostPort, serviceName string) {

	// set up a span reporter
	reporter := zipkinhttp.NewReporter(zipkinHTTPEndpoint)
	//defer reporter.Close()

	// create our local service endpoint
	endpoint, err := zipkin.NewEndpoint(serviceName, hostPort)
	if err != nil {
		log.Fatalf("unable to create local endpoint: %+v\n", err)
	}

	// initialize our tracer
	nativeTracer, err := zipkin.NewTracer(reporter, zipkin.WithLocalEndpoint(endpoint))
	if err != nil {
		log.Fatalf("unable to create tracer: %+v\n", err)
	}

	// use zipkin-go-opentracing to wrap our tracer
	tracer := zipkinot.Wrap(nativeTracer)

	// optionally set as Global OpenTracing tracer instance
	opentracing.SetGlobalTracer(tracer)
}

//ZipkinClientFilter gets tars client filter for zipkin opentracing.
func ZipkinClientFilter() tars.ClientFilter {
	return func(ctx context.Context, msg *tars.Message, invoke tars.Invoke, timeout time.Duration) (err error) {
		var pCtx opentracing.SpanContext
		req := msg.Req
		if parent := opentracing.SpanFromContext(ctx); parent != nil {
			pCtx = parent.Context()
		}
		cSpan := opentracing.GlobalTracer().StartSpan(
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
		if req.Status != nil {
			err = opentracing.GlobalTracer().Inject(cSpan.Context(), opentracing.TextMap, opentracing.TextMapCarrier(req.Status))
			if err != nil {
				logger.Error("inject span to status error:", err)
			}
		} else {
			s := make(map[string]string)
			err = opentracing.GlobalTracer().Inject(cSpan.Context(), opentracing.TextMap, opentracing.TextMapCarrier(s))
			if err != nil {
				logger.Error("inject span to status error:", err)
			} else {
				req.Status = s
			}
		}
		err = invoke(ctx, msg, timeout)
		if err != nil {
			ext.Error.Set(cSpan, true)
			cSpan.LogFields(oplog.String("event", "error"), oplog.String("message", err.Error()))
		}

		return err
	}
}

//ZipkinServerFilter gets tars server filter for zipkin opentraicng.
func ZipkinServerFilter() tars.ServerFilter {
	return func(ctx context.Context, d tars.Dispatch, f interface{},
		req *requestf.RequestPacket, resp *requestf.ResponsePacket, withContext bool) (err error) {

		pCtx, err := opentracing.GlobalTracer().Extract(opentracing.TextMap, opentracing.TextMapCarrier(req.Status))
		if err != nil && err != opentracing.ErrSpanContextNotFound {
			return err
		}
		serverSpan := opentracing.GlobalTracer().StartSpan(
			req.SFuncName,
			ext.RPCServerOption(pCtx),
			ext.SpanKindRPCServer,
		)
		defer serverSpan.Finish()
		cfg := tars.GetServerConfig()
		serverSpan.SetTag("server.ipv4", cfg.LocalIP)
		serverSpan.SetTag("server.port", strconv.Itoa(int(cfg.Adapters[req.SServantName+"Adapter"].Endpoint.Port)))
		if cfg.Enableset {
			serverSpan.SetTag("tars.set_division", cfg.Setdivision)
		}
		ctx = opentracing.ContextWithSpan(ctx, serverSpan)
		err = d(ctx, f, req, resp, withContext)
		if err != nil {
			ext.Error.Set(serverSpan, true)
			serverSpan.LogFields(oplog.String("event", "error"), oplog.String("message", err.Error()))
		}
		return err

	}
}
