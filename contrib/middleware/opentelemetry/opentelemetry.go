package opentelemetry

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/TarsCloud/TarsGo/tars"
	"github.com/TarsCloud/TarsGo/tars/protocol/res/requestf"
	"github.com/TarsCloud/TarsGo/tars/util/current"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

const (
	instrumentationName = "github.com/TarsCloud/TarsGo/tars/middleware/opentelemetry"
	TarsRpcRetKey       = attribute.Key("tars.rpc.ret")
)

type Opentelemetry struct {
	Propagators    propagation.TextMapPropagator
	TracerProvider trace.TracerProvider
	MeterProvider  metric.MeterProvider
}

type Option func(*Opentelemetry)

func WithTracerProvider(tp trace.TracerProvider) Option {
	return func(o *Opentelemetry) {
		o.TracerProvider = tp
	}
}

func WithPropagators(p propagation.TextMapPropagator) Option {
	return func(o *Opentelemetry) {
		o.Propagators = p
	}
}

func New(opts ...Option) *Opentelemetry {
	o := &Opentelemetry{
		TracerProvider: otel.GetTracerProvider(),
		Propagators:    otel.GetTextMapPropagator(),
	}
	for _, opt := range opts {
		opt(o)
	}
	return o
}

func (o *Opentelemetry) BuildServerFilter() tars.ServerFilterMiddleware {
	localIp := getOutboundIP()
	tracer := o.TracerProvider.Tracer(instrumentationName)
	return func(next tars.ServerFilter) tars.ServerFilter {
		return func(ctx context.Context, d tars.Dispatch, f interface{}, req *requestf.RequestPacket, resp *requestf.ResponsePacket, withContext bool) (err error) {
			ip, _ := current.GetClientIPFromContext(ctx)
			port, _ := current.GetClientPortFromContext(ctx)
			var span trace.Span
			ctx = o.extract(ctx, req)
			servants := strings.Split(req.SServantName, ".")
			ctx, span = tracer.Start(
				ctx,
				fmt.Sprintf("%s.%s", servants[2], req.SFuncName),
				trace.WithSpanKind(trace.SpanKindServer),
				trace.WithAttributes(
					attribute.String("tars.interface", req.SServantName),
					attribute.String("tars.method", req.SFuncName),
					attribute.String("tars.local_ip", localIp),
					attribute.Int("tars.request_id", int(req.IRequestId)),
					attribute.String("tars.client.ipv4", ip),
					attribute.String("tars.client.port", port),
					attribute.String("tars.server.version", tars.Version),
				),
			)
			defer span.End()
			cfg := tars.GetServerConfig()
			if cfg.Enableset {
				span.SetAttributes(attribute.String("tars.set_division", cfg.Setdivision))
			}
			err = next(ctx, d, f, req, resp, withContext)
			if err != nil {
				span.SetStatus(codes.Error, "server failed")
				span.RecordError(err)
				span.SetAttributes(TarsRpcRetKey.Int64(int64(codes.Error)))
			} else {
				span.SetAttributes(TarsRpcRetKey.Int64(int64(resp.IRet)))
			}
			return err
		}
	}
}

func (o *Opentelemetry) extract(ctx context.Context, req *requestf.RequestPacket) context.Context {
	if req.Status == nil {
		req.Status = make(map[string]string)
	}
	return o.Propagators.Extract(ctx, propagation.MapCarrier(req.Status))
}

func (o *Opentelemetry) BuildHttpHandler() func(next http.Handler) http.Handler {
	tracer := o.TracerProvider.Tracer(instrumentationName)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var span trace.Span
			reqCtx := r.Context()
			reqCtx = o.Propagators.Extract(reqCtx, propagation.HeaderCarrier(r.Header))
			reqCtx, span = tracer.Start(
				reqCtx,
				r.URL.Path,
				trace.WithSpanKind(trace.SpanKindServer),
				trace.WithAttributes(
					attribute.String("http.method", r.Method),
					attribute.String("http.url", r.URL.String()),
					attribute.String("http.scheme", r.URL.Scheme),
					attribute.String("http.proto", r.Proto),
					attribute.String("peer.hostname", r.Host),
					attribute.String("peer.address", r.RemoteAddr),
					attribute.String("component", "web"),
				),
			)
			defer span.End()

			r = r.WithContext(reqCtx)
			recorder := httptest.NewRecorder()
			next.ServeHTTP(recorder, r)
			span.SetAttributes(attribute.Int("http.status", recorder.Code))
			for k, v := range recorder.Result().Header {
				w.Header()[k] = v
			}
			w.WriteHeader(recorder.Code)
			_, err := w.Write(recorder.Body.Bytes())
			if err != nil {
				span.SetStatus(codes.Error, "http server write failed")
				span.RecordError(err)
			}
		})
	}
}

func (o *Opentelemetry) BuildClientFilter() tars.ClientFilterMiddleware {
	localIp := getOutboundIP()
	tracer := o.TracerProvider.Tracer(instrumentationName)
	return func(next tars.ClientFilter) tars.ClientFilter {
		return func(ctx context.Context, msg *tars.Message, invoke tars.Invoke, timeout time.Duration) (err error) {
			var span trace.Span
			servants := strings.Split(msg.Req.SServantName, ".")
			ctx, span = tracer.Start(
				ctx,
				fmt.Sprintf("%s.%s", servants[2], msg.Req.SFuncName),
				trace.WithSpanKind(trace.SpanKindClient),
				trace.WithAttributes(
					attribute.String("tars.interface", msg.Req.SServantName),
					attribute.String("tars.method", msg.Req.SFuncName),
					attribute.Int("tars.request_id", int(msg.Req.IRequestId)),
					attribute.String("tars.local_ip", localIp),
					attribute.String("tars.protocol", "tars"),
					attribute.String("tars.client.version", tars.Version),
				),
			)
			ctx = o.inject(ctx, msg)
			defer func() {
				ip, _ := current.GetServerIPFromContext(ctx)
				port, _ := current.GetServerPortFromContext(ctx)
				span.SetAttributes(attribute.String("tars.server.ipv4", ip))
				span.SetAttributes(attribute.String("tars.server.port", port))
				span.End()
			}()

			err = next(ctx, msg, invoke, timeout)
			if err != nil {
				span.SetStatus(codes.Error, "client failed")
				span.RecordError(err)
				span.SetAttributes(TarsRpcRetKey.Int64(int64(codes.Error)))
			} else {
				span.SetAttributes(TarsRpcRetKey.Int64(int64(msg.Resp.IRet)))
			}
			return err
		}
	}
}

func (o *Opentelemetry) inject(ctx context.Context, msg *tars.Message) context.Context {
	if msg.Req.Status == nil {
		msg.Req.Status = make(map[string]string)
	}
	o.Propagators.Inject(ctx, propagation.MapCarrier(msg.Req.Status))
	return ctx
}

func getOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return ""
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}
