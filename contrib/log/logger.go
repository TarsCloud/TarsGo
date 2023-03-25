package log

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/TarsCloud/TarsGo/tars/util/current"
	"github.com/TarsCloud/TarsGo/tars/util/rogger"
	"go.opentelemetry.io/otel/trace"
)

// Logger is the struct with name and writer.
type Logger struct {
	*rogger.Logger
}

var (
	loggerMutex sync.Mutex
	loggerMap   = make(map[string]*Logger)

	callerSkip = 3
)

// GetCtxLogger return an logger instance
func GetCtxLogger(name string) *Logger {
	loggerMutex.Lock()
	defer loggerMutex.Unlock()
	if lg, ok := loggerMap[name]; ok {
		return lg
	}
	lg := &Logger{
		Logger: rogger.GetLogger(name),
	}
	loggerMap[name] = lg
	return lg
}

func (l *Logger) Flush(_ context.Context) {
	rogger.FlushLogger()
}

// Debug logs interface in debug loglevel.
func (l *Logger) Debug(ctx context.Context, v ...interface{}) {
	l.Writef(ctx, 0, rogger.DEBUG, "", v)
}

// Info logs interface in Info loglevel.
func (l *Logger) Info(ctx context.Context, v ...interface{}) {
	l.Writef(ctx, 0, rogger.INFO, "", v)
}

// Warn logs interface in warning loglevel
func (l *Logger) Warn(ctx context.Context, v ...interface{}) {
	l.Writef(ctx, 0, rogger.WARN, "", v)
}

// Error logs interface in Error loglevel
func (l *Logger) Error(ctx context.Context, v ...interface{}) {
	l.Writef(ctx, 0, rogger.ERROR, "", v)
}

// Debugf logs interface in debug loglevel with formatting string
func (l *Logger) Debugf(ctx context.Context, format string, v ...interface{}) {
	l.Writef(ctx, 0, rogger.DEBUG, format, v)
}

// Infof logs interface in Infof loglevel with formatting string
func (l *Logger) Infof(ctx context.Context, format string, v ...interface{}) {
	l.Writef(ctx, 0, rogger.INFO, format, v)
}

// Warnf logs interface in warning loglevel with formatting string
func (l *Logger) Warnf(ctx context.Context, format string, v ...interface{}) {
	l.Writef(ctx, 0, rogger.WARN, format, v)
}

// Errorf logs interface in Error loglevel with formatting string
func (l *Logger) Errorf(ctx context.Context, format string, v ...interface{}) {
	l.Writef(ctx, 0, rogger.ERROR, format, v)
}

func (l *Logger) Writef(ctx context.Context, depth int, level rogger.LogLevel, format string, v []interface{}) {
	if level < rogger.GetLogLevel() {
		return
	}

	if rogger.GetLogFormat() == rogger.Json {
		l.WriteLog(l.writeJson(ctx, depth, level, format, v))
	} else {
		l.WriteLog(l.writeLine(ctx, depth, level, format, v))
	}
}

func (l *Logger) writeLine(ctx context.Context, depth int, level rogger.LogLevel, format string, v []interface{}) []byte {
	buf := bytes.NewBuffer(nil)
	if l.Writer().NeedPrefix() {
		fmt.Fprintf(buf, "%s|", time.Now().Format("2006-01-02 15:04:05.000"))
		if len(l.Prefix()) > 0 {
			fmt.Fprintf(buf, "%s|", l.Prefix())
		}
		// trace
		span := trace.SpanFromContext(ctx)
		if span.IsRecording() {
			sc := span.SpanContext()
			fmt.Fprintf(buf, "%s|%s|", sc.TraceID(), sc.SpanID())
		} else if t, ok := current.GetTarsTrace(ctx); ok {
			sc := t.SpanContext()
			fmt.Fprintf(buf, "%s|%s|", sc.TraceID(), sc.SpanID())
		}

		if rogger.CallerFlag() {
			pc, file, line, ok := runtime.Caller(depth + callerSkip)
			if !ok {
				file = "???"
				line = 0
			} else {
				file = filepath.Base(file)
			}
			fmt.Fprintf(buf, "%s:%s:%d|", file, rogger.FuncName(runtime.FuncForPC(pc)), line)
		}
		if rogger.IsColored() && l.IsConsoleWriter() {
			buf.WriteString(level.ColoredString())
		} else {
			buf.WriteString(level.String())
		}
		buf.WriteByte('|')
	}

	if format == "" {
		fmt.Fprint(buf, v...)
	} else {
		fmt.Fprintf(buf, format, v...)
	}
	if l.Writer().NeedPrefix() {
		buf.WriteByte('\n')
	}
	return buf.Bytes()
}

func (l *Logger) writeJson(ctx context.Context, depth int, level rogger.LogLevel, format string, v []interface{}) []byte {
	log := rogger.JsonLog{}
	log.Pre = l.Prefix()
	log.Time = time.Now().Format("2006-01-02 15:04:05.000")
	// trace
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		sc := span.SpanContext()
		log.TraceId = sc.TraceID().String()
		log.SpanId = sc.SpanID().String()
	} else if t, ok := current.GetTarsTrace(ctx); ok {
		sc := t.SpanContext()
		log.TraceId = sc.TraceID()
		log.SpanId = sc.SpanID()
	}

	if rogger.CallerFlag() {
		pc, file, line, ok := runtime.Caller(depth + callerSkip)
		if !ok {
			file = "???"
			line = 0
		} else {
			file = filepath.Base(file)
		}
		log.Func = rogger.FuncName(runtime.FuncForPC(pc))
		log.File = fmt.Sprintf("%s:%d", file, line)
	}
	log.Level = level.String()

	if format == "" {
		log.Msg = fmt.Sprint(v...)
	} else {
		log.Msg = fmt.Sprintf(format, v...)
	}
	buf := &bytes.Buffer{}
	encoder := json.NewEncoder(buf)
	encoder.SetEscapeHTML(false)
	_ = encoder.Encode(log)
	return buf.Bytes()
}
