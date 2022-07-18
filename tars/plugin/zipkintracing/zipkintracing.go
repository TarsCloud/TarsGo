package zipkintracing

import (
	"context"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"time"

	"github.com/TarsCloud/TarsGo/tars"
	"github.com/TarsCloud/TarsGo/tars/protocol/res/requestf"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	oplog "github.com/opentracing/opentracing-go/log"
	zipkinot "github.com/openzipkin-contrib/zipkin-go-opentracing"
	"github.com/openzipkin/zipkin-go"
	"github.com/openzipkin/zipkin-go/proto/zipkin_proto3"
	zipkinreporter "github.com/openzipkin/zipkin-go/reporter"
	zipkinhttp "github.com/openzipkin/zipkin-go/reporter/http"
	zipkinkafka "github.com/openzipkin/zipkin-go/reporter/kafka"
)

var logger = tars.GetLogger("tracing")
var tracerMap = map[string]opentracing.Tracer{}
var isTrace = false

// Init is used to init opentracing and zipkin
func Init(zipkinHTTPEndpoint string, sameSpan bool, traceID128Bit bool, debug bool,
	hostPort, serviceName string) {
	// set up a span reporter
	reporter := zipkinhttp.NewReporter(zipkinHTTPEndpoint)
	// defer reporter.Close()

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
	isTrace = true
}

// InitV2 is used to init opentracing and zipkin, all configs are loaded from server config
func InitV2() {
	serverConfig := tars.GetServerConfig()
	isTrace = serverConfig.SampleRate > 0

	if !isTrace {
		return
	}

	var (
		rpt        zipkinreporter.Reporter
		serializer zipkinreporter.SpanSerializer
		err        error
	)
	switch serverConfig.SampleEncoding {
	case "json":
		serializer = zipkinreporter.JSONSerializer{}
	case "proto":
		serializer = zipkin_proto3.SpanSerializer{}
	default:
		log.Fatalf("unsupported sample encoding: %s\n", serverConfig.SampleEncoding)
	}

	switch serverConfig.SampleType {
	case "http":
		url := strings.TrimRight(serverConfig.SampleAddress, "/") + "/api/v2/spans"
		rpt = zipkinhttp.NewReporter(url, zipkinhttp.Serializer(serializer))
	case "kafka":
		rpt, err = zipkinkafka.NewReporter(
			strings.Split(serverConfig.SampleAddress, ","), zipkinkafka.Serializer(serializer),
		)
		if err != nil {
			log.Fatalf("unable to create tracer: %+v\n", err)
		}
	default:
		log.Fatalf("unsupported sample type: %s\n", serverConfig.SampleType)
	}

	sampler, err := zipkin.NewCountingSampler(serverConfig.SampleRate)
	if err != nil {
		log.Fatalf("unable to create sampler: %+v\n", err)
	}

	for _, config := range serverConfig.Adapters {
		endpoint, err := zipkin.NewEndpoint(
			config.Obj, config.Endpoint.Host+":"+strconv.FormatInt(int64(config.Endpoint.Port), 10),
		)

		if err != nil {
			log.Fatalf("unable to create local endpoint: %+v\n", err)
		}

		nativeTracer, err := zipkin.NewTracer(rpt, zipkin.WithLocalEndpoint(endpoint), zipkin.WithSampler(sampler))
		if err != nil {
			log.Fatalf("unable to create tracer: %+v\n", err)
		}

		// use zipkin-go-opentracing to wrap our tracer
		tracer := zipkinot.Wrap(nativeTracer)
		tracerMap[config.Obj] = tracer
	}

	// If the request is not called by any servant(such as job, queue, scheduler), use opentracing.GlobalTracer()
	endpoint, err := zipkin.NewEndpoint(serverConfig.App+"."+serverConfig.Server, "")
	if err != nil {
		log.Fatalf("unable to create local endpoint: %+v\n", err)
	}

	nativeTracer, err := zipkin.NewTracer(rpt, zipkin.WithLocalEndpoint(endpoint), zipkin.WithSampler(sampler))
	if err != nil {
		log.Fatalf("unable to create tracer: %+v\n", err)
	}

	opentracing.SetGlobalTracer(zipkinot.Wrap(nativeTracer))
}

type contextKey struct{}

var servantName contextKey

// ContextWithServant add servant to context
func ContextWithServant(ctx context.Context, servant string) context.Context {
	return context.WithValue(ctx, servantName, servant)
}

// ServantFromContext gets servant from context
func ServantFromContext(ctx context.Context) string {
	name := ctx.Value(servantName)
	if name == nil {
		return ""
	}

	return name.(string)
}

// GetTracer gets tracer with the servant name
func GetTracer(servant string) opentracing.Tracer {
	return tracerMap[servant]
}

// ZipkinClientFilter gets tars client filter for zipkin opentracing.
func ZipkinClientFilter() tars.ClientFilter {
	return func(ctx context.Context, msg *tars.Message, invoke tars.Invoke, timeout time.Duration) (err error) {
		if !isTrace {
			return invoke(ctx, msg, timeout)
		}

		var pCtx opentracing.SpanContext
		if parent := opentracing.SpanFromContext(ctx); parent != nil {
			pCtx = parent.Context()
		}

		cfg := tars.GetServerConfig()
		var tracer opentracing.Tracer
		var port string
		if servant := ServantFromContext(ctx); servant != "" {
			tracer = GetTracer(servant)
			port = strconv.Itoa(int(cfg.Adapters[servant+"Adapter"].Endpoint.Port))
		}

		if tracer == nil {
			tracer = opentracing.GlobalTracer()
		}

		cSpan := tracer.StartSpan(
			msg.Req.SFuncName,
			opentracing.ChildOf(pCtx),
			ext.SpanKindRPCClient,
		)

		defer cSpan.Finish()
		cSpan.SetTag("client.ipv4", cfg.LocalIP)
		cSpan.SetTag("client.port", port)
		cSpan.SetTag("tars.interface", msg.Req.SServantName)
		cSpan.SetTag("tars.method", msg.Req.SFuncName)
		cSpan.SetTag("tars.protocol", "tars")
		cSpan.SetTag("tars.client.version", tars.Version)
		if msg.Req.Status != nil {
			err = tracer.Inject(cSpan.Context(), opentracing.TextMap, opentracing.TextMapCarrier(msg.Req.Status))
			if err != nil {
				logger.Error("inject span to status error:", err)
			}
		} else {
			s := make(map[string]string)
			err = tracer.Inject(cSpan.Context(), opentracing.TextMap, opentracing.TextMapCarrier(s))
			if err != nil {
				logger.Error("inject span to status error:", err)
			} else {
				msg.Req.Status = s
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

// ZipkinServerFilter gets tars server filter for zipkin opentracing.
func ZipkinServerFilter() tars.ServerFilter {
	return func(ctx context.Context, d tars.Dispatch, f interface{},
		req *requestf.RequestPacket, resp *requestf.ResponsePacket, withContext bool) (err error) {
		if !isTrace {
			return d(ctx, f, req, resp, withContext)
		}
		tracer := GetTracer(req.SServantName)
		if tracer == nil {
			tracer = opentracing.GlobalTracer()
		}
		ctx = ContextWithServant(ctx, req.SServantName)
		var serverSpan opentracing.Span
		pCtx, err := tracer.Extract(opentracing.TextMap, opentracing.TextMapCarrier(req.Status))
		if err == nil {
			serverSpan = tracer.StartSpan(req.SFuncName, ext.RPCServerOption(pCtx), ext.SpanKindRPCServer)
		} else {
			serverSpan = tracer.StartSpan(req.SFuncName, ext.SpanKindRPCServer)
		}

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

// ZipkinHttpMiddleware zipkin http server router middleware
func ZipkinHttpMiddleware(next http.Handler) http.Handler {
	servantMap := make(map[string]string)
	serverConfig := tars.GetServerConfig()
	for _, adapterConfig := range serverConfig.Adapters {
		servantMap[adapterConfig.Endpoint.Host+":"+strconv.FormatInt(
			int64(adapterConfig.Endpoint.Port), 10,
		)] = adapterConfig.Obj
	}
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			addr, ok := r.Context().Value(http.LocalAddrContextKey).(net.Addr)
			if !ok {
				next.ServeHTTP(w, r)
				return
			}

			servant := servantMap[addr.String()]
			tracer := GetTracer(servant)
			if tracer == nil {
				next.ServeHTTP(w, r)
				return
			}

			var serverSpan opentracing.Span
			pCtx, err := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
			if err == nil {
				serverSpan = tracer.StartSpan(r.URL.Path, ext.RPCServerOption(pCtx), ext.SpanKindRPCServer)
			} else {
				serverSpan = tracer.StartSpan(r.URL.Path, ext.SpanKindRPCServer)
			}

			defer serverSpan.Finish()
			cfg := tars.GetServerConfig()
			ext.HTTPUrl.Set(serverSpan, r.RequestURI)
			ext.HTTPMethod.Set(serverSpan, r.Method)
			serverSpan.SetTag("server.ipv4", cfg.LocalIP)
			serverSpan.SetTag("server.port", strconv.Itoa(int(cfg.Adapters[servant+"Adapter"].Endpoint.Port)))
			if cfg.Enableset {
				serverSpan.SetTag("tars.set_division", cfg.Setdivision)
			}
			ctx := opentracing.ContextWithSpan(r.Context(), serverSpan)
			ctx = ContextWithServant(ctx, servant)
			r = r.WithContext(ctx)
			recorder := httptest.NewRecorder()
			next.ServeHTTP(recorder, r)
			for k, v := range recorder.Result().Header {
				w.Header()[k] = v
			}
			w.WriteHeader(recorder.Code)
			_, _ = w.Write(recorder.Body.Bytes())

			ext.HTTPStatusCode.Set(serverSpan, uint16(recorder.Code))
			if err != nil {
				ext.Error.Set(serverSpan, true)
				serverSpan.LogFields(oplog.String("event", "error"), oplog.String("message", err.Error()))
			}
		},
	)
}
