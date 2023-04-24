package zipkintracing

import (
	"context"
	"fmt"
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

// Init is used to init opentracing and zipkin, all configs are loaded from server config
// /tars/application/server add the following config
// samplerate=0.5
// sampleaddress=http://127.0.0.1:9411
// sampletype=http
// sampleencoding=json
func Init() {
	cfg := tars.GetServerConfig()
	isTrace = cfg.SampleRate > 0
	if !isTrace {
		return
	}

	var (
		rpt        zipkinreporter.Reporter
		serializer zipkinreporter.SpanSerializer
		err        error
	)
	switch cfg.SampleEncoding {
	case "json":
		serializer = zipkinreporter.JSONSerializer{}
	case "proto":
		serializer = zipkin_proto3.SpanSerializer{}
	default:
		log.Fatalf("unsupported sample encoding: %s\n", cfg.SampleEncoding)
	}

	switch cfg.SampleType {
	case "http":
		url := strings.TrimRight(cfg.SampleAddress, "/") + "/api/v2/spans"
		rpt = zipkinhttp.NewReporter(url, zipkinhttp.Serializer(serializer))
	case "kafka":
		brokers := strings.Split(cfg.SampleAddress, ",")
		rpt, err = zipkinkafka.NewReporter(brokers, zipkinkafka.Serializer(serializer))
		if err != nil {
			log.Fatalf("unable to create tracer: %v\n", err)
		}
	default:
		log.Fatalf("unsupported sample type: %s\n", cfg.SampleType)
	}

	sampler, err := zipkin.NewCountingSampler(cfg.SampleRate)
	if err != nil {
		log.Fatalf("unable to create sampler: %v\n", err)
	}

	for _, adapter := range cfg.Adapters {
		endpoint, err := zipkin.NewEndpoint(adapter.Obj, fmt.Sprintf("%s:%d", adapter.Endpoint.Host, adapter.Endpoint.Port))
		if err != nil {
			log.Fatalf("unable to create local endpoint: %v\n", err)
		}

		nativeTracer, err := zipkin.NewTracer(rpt, zipkin.WithLocalEndpoint(endpoint), zipkin.WithSampler(sampler))
		if err != nil {
			log.Fatalf("unable to create tracer: %v\n", err)
		}

		// use zipkin-go-opentracing to wrap our tracer
		tracerMap[adapter.Obj] = zipkinot.Wrap(nativeTracer)
	}

	// If the request is not called by any servant(such as job, queue, scheduler), use opentracing.GlobalTracer()
	endpoint, err := zipkin.NewEndpoint(fmt.Sprintf("%s.%s", cfg.App, cfg.Server), "")
	if err != nil {
		log.Fatalf("unable to create local endpoint: %v\n", err)
	}

	nativeTracer, err := zipkin.NewTracer(rpt, zipkin.WithLocalEndpoint(endpoint), zipkin.WithSampler(sampler))
	if err != nil {
		log.Fatalf("unable to create tracer: %v\n", err)
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
func ZipkinClientFilter() tars.ClientFilterMiddleware {
	return func(next tars.ClientFilter) tars.ClientFilter {
		return func(ctx context.Context, msg *tars.Message, invoke tars.Invoke, timeout time.Duration) (err error) {
			if !isTrace {
				return next(ctx, msg, invoke, timeout)
			}

			var spanCtx opentracing.SpanContext
			if parent := opentracing.SpanFromContext(ctx); parent != nil {
				spanCtx = parent.Context()
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

			span := tracer.StartSpan(msg.Req.SFuncName, opentracing.ChildOf(spanCtx), ext.SpanKindRPCClient)
			defer span.Finish()
			span.SetTag("client.ipv4", cfg.LocalIP)
			span.SetTag("client.port", port)
			span.SetTag("tars.interface", msg.Req.SServantName)
			span.SetTag("tars.method", msg.Req.SFuncName)
			span.SetTag("tars.protocol", "tars")
			span.SetTag("tars.client.version", tars.Version)
			if msg.Req.Status == nil {
				msg.Req.Status = make(map[string]string)
			}
			err = tracer.Inject(span.Context(), opentracing.TextMap, opentracing.TextMapCarrier(msg.Req.Status))
			if err != nil {
				logger.Error("inject span to status error:", err)
			}
			err = next(ctx, msg, invoke, timeout)
			if err != nil {
				ext.Error.Set(span, true)
				span.LogFields(oplog.String("event", "error"), oplog.String("message", err.Error()))
			}
			return err
		}
	}
}

// ZipkinServerFilter gets tars server filter for zipkin opentracing.
func ZipkinServerFilter() tars.ServerFilterMiddleware {
	return func(next tars.ServerFilter) tars.ServerFilter {
		return func(ctx context.Context, d tars.Dispatch, f interface{}, req *requestf.RequestPacket, resp *requestf.ResponsePacket, withContext bool) (err error) {
			if !isTrace {
				return next(ctx, d, f, req, resp, withContext)
			}
			tracer := GetTracer(req.SServantName)
			if tracer == nil {
				tracer = opentracing.GlobalTracer()
			}
			ctx = ContextWithServant(ctx, req.SServantName)
			var span opentracing.Span
			spanCtx, err := tracer.Extract(opentracing.TextMap, opentracing.TextMapCarrier(req.Status))
			if err == nil {
				span = tracer.StartSpan(req.SFuncName, ext.RPCServerOption(spanCtx), ext.SpanKindRPCServer)
			} else {
				span = tracer.StartSpan(req.SFuncName, ext.SpanKindRPCServer)
			}

			defer span.Finish()
			cfg := tars.GetServerConfig()
			span.SetTag("server.ipv4", cfg.LocalIP)
			span.SetTag("server.port", strconv.Itoa(int(cfg.Adapters[req.SServantName+"Adapter"].Endpoint.Port)))
			if cfg.Enableset {
				span.SetTag("tars.set_division", cfg.Setdivision)
			}
			ctx = opentracing.ContextWithSpan(ctx, span)
			err = next(ctx, d, f, req, resp, withContext)
			if err != nil {
				ext.Error.Set(span, true)
				span.LogFields(oplog.String("event", "error"), oplog.String("message", err.Error()))
			}
			return err
		}
	}
}

// ZipkinHttpMiddleware zipkin http server router middleware
func ZipkinHttpMiddleware(next http.Handler) http.Handler {
	cfg := tars.GetServerConfig()
	servantMap := make(map[string]string)
	for _, adapter := range cfg.Adapters {
		servantMap[fmt.Sprintf("%s:%d", adapter.Endpoint.Host, adapter.Endpoint.Port)] = adapter.Obj
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

		var span opentracing.Span
		spanCtx, err := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
		if err == nil {
			span = tracer.StartSpan(r.URL.Path, ext.RPCServerOption(spanCtx), ext.SpanKindRPCServer)
		} else {
			span = tracer.StartSpan(r.URL.Path, ext.SpanKindRPCServer)
		}

		defer span.Finish()
		ext.HTTPUrl.Set(span, r.RequestURI)
		ext.HTTPMethod.Set(span, r.Method)
		span.SetTag("server.ipv4", cfg.LocalIP)
		span.SetTag("server.port", strconv.Itoa(int(cfg.Adapters[servant+"Adapter"].Endpoint.Port)))
		if cfg.Enableset {
			span.SetTag("tars.set_division", cfg.Setdivision)
		}
		ctx := opentracing.ContextWithSpan(r.Context(), span)
		ctx = ContextWithServant(ctx, servant)
		r = r.WithContext(ctx)
		recorder := httptest.NewRecorder()
		next.ServeHTTP(recorder, r)
		for k, v := range recorder.Result().Header {
			w.Header()[k] = v
		}
		w.WriteHeader(recorder.Code)
		_, err = w.Write(recorder.Body.Bytes())

		ext.HTTPStatusCode.Set(span, uint16(recorder.Code))
		if err != nil {
			ext.Error.Set(span, true)
			span.LogFields(oplog.String("event", "error"), oplog.String("message", err.Error()))
		}
	})
}
