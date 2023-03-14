package log

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/TarsCloud/TarsGo/tars/util/current"
	"github.com/TarsCloud/TarsGo/tars/util/rogger"

	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

const (
	traceIDStr = "4bf92f3577b34da6a3ce929d0e0e4736"
	spanIDStr  = "00f067aa0ba902b7"
)

var (
	traceID = mustTraceIDFromHex(traceIDStr)
	spanID  = mustSpanIDFromHex(spanIDStr)
)

func mustTraceIDFromHex(s string) (t trace.TraceID) {
	var err error
	t, err = trace.TraceIDFromHex(s)
	if err != nil {
		panic(err)
	}
	return
}

func mustSpanIDFromHex(s string) (t trace.SpanID) {
	var err error
	t, err = trace.SpanIDFromHex(s)
	if err != nil {
		panic(err)
	}
	return
}

func TestLog(t *testing.T) {
	rogger.SetLevel(rogger.DEBUG)
	ctx := context.Background()
	l := GetCtxLogger("log")
	defer l.Flush(ctx)
	l.SetPrefix("prefix")
	testCases := []struct {
		name string
		ctx  func(t *testing.T) context.Context
	}{
		{
			name: "background",
			ctx: func(t *testing.T) context.Context {
				return context.Background()
			},
		},
		{
			name: "opentelemetry",
			ctx: func(t *testing.T) context.Context {
				exporter, err := stdouttrace.New(
					// Use human-readable output.
					stdouttrace.WithPrettyPrint(),
					// Do not print timestamps for the demo.
					stdouttrace.WithoutTimestamps(),
				)
				require.NoError(t, err)
				tp := sdktrace.NewTracerProvider(sdktrace.WithBatcher(exporter))
				otelCtx, _ := tp.Tracer("opentelemetry").Start(context.Background(), "span")
				return otelCtx
			},
		},
		{
			name: "tars trace",
			ctx: func(t *testing.T) context.Context {
				tarsCtx := current.ContextWithTarsCurrent(context.Background())
				current.OpenTarsTrace(tarsCtx, true)
				tt, _ := current.GetTarsTrace(tarsCtx)
				tt.NewSpan()
				return tarsCtx
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for _, format := range []rogger.LogFormat{rogger.Text, rogger.Json} {
				rogger.SetFormat(format)
				l.Debug(tc.ctx(t), "DEBUG LOG")
				l.Debugf(tc.ctx(t), "DEBUG LOG FORMAT: %s", format.String())
				l.Info(tc.ctx(t), "INfO LOG")
				l.Infof(tc.ctx(t), "INfO LOG FORMAT: %s", format.String())
				l.Warn(tc.ctx(t), "WARN LOG")
				l.Warnf(tc.ctx(t), "WARN LOG FORMAT: %s", format.String())
				l.Error(tc.ctx(t), "ERROR LOG")
				l.Errorf(tc.ctx(t), "ERROR LOG FORMAT: %s", format.String())
			}
		})
	}
}
