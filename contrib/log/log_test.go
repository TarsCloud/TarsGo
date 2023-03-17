package log

import (
	"context"
	"testing"

	"github.com/TarsCloud/TarsGo/tars/util/rogger"
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
	l := getCtxLogger("log")
	l.SetPrefix("prefix")
	ctx := context.Background()
	defer l.Flush(ctx)
	ctx = trace.ContextWithSpanContext(ctx, trace.NewSpanContext(trace.SpanContextConfig{
		TraceID: traceID,
		SpanID:  spanID,
	}))
	for _, format := range []rogger.LogFormat{rogger.Text, rogger.Json} {
		rogger.SetFormat(format)
		l.Debug(ctx, "DEBUG LOG")
		l.Debugf(ctx, "DEBUG LOG FORMAT: %s", format.String())
		l.Info(ctx, "INfO LOG")
		l.Infof(ctx, "INfO LOG FORMAT: %s", format.String())
		l.Warn(ctx, "WARN LOG")
		l.Warnf(ctx, "WARN LOG FORMAT: %s", format.String())
		l.Error(ctx, "ERROR LOG")
		l.Errorf(ctx, "ERROR LOG FORMAT: %s", format.String())
	}
}
