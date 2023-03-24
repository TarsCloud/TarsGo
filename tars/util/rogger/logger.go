package rogger

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"time"
)

// DEBUG loglevel
const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	OFF
)

// Format LogFormat
const (
	Text LogFormat = iota
	Json
)

var (
	logLevel  = DEBUG
	logFormat = Text
	colored   = false

	logQueue    = make(chan *logValue, 10000)
	loggerMutex sync.Mutex
	loggerMap   = make(map[string]*Logger)
	//writeDone   = make(chan bool)
	callerSkip = 3
	callerFlag = true

	waitFlushTimeout       = time.Second
	syncDone, syncCancel   = context.WithCancel(context.Background())
	asyncDone, asyncCancel = context.WithCancel(context.Background())
)

// Logger is the struct with name and writer.
type Logger struct {
	name   string
	prefix string
	writer LogWriter
}

type JsonLog struct {
	Pre     string `json:"pre,omitempty"`
	Time    string `json:"time"`
	TraceId string `json:"trace_id,omitempty"`
	SpanId  string `json:"span_id,omitempty"`
	Func    string `json:"func"`
	File    string `json:"file"`
	Level   string `json:"level"`
	Msg     string `json:"msg"`
}

// LogLevel is uint8 type
type LogLevel uint8

// LogFormat is uint8 format
type LogFormat uint8

type logValue struct {
	// level  LogLevel
	// fileNo string
	value  []byte
	writer LogWriter
}

func init() {
	go flushLog()
}

// String returns the LogLevel to string.
func (lv *LogLevel) String() string {
	switch *lv {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// ColoredString enable colored level string when use console writer
func (lv *LogLevel) ColoredString() string {
	switch *lv {
	case DEBUG:
		return "\x1b[34mDEBUG\x1b[0m" // blue
	case INFO:
		return "\x1b[32mINFO\x1b[0m" //green
	case WARN:
		return "\x1b[33mWARN\x1b[0m" // yellow
	case ERROR:
		return "\x1b[31mERROR\x1b[0m" //cred
	default:
		return "\x1b[37mUNKNOWN\x1b[0m" // white
	}
}

// String returns the LogFormat to string.
func (lv *LogFormat) String() string {
	switch *lv {
	case Text:
		return "LINE"
	case Json:
		return "JSON"
	default:
		return "UNKNOWN"
	}
}

// GetLogger return an logger instance
func GetLogger(name string) *Logger {
	loggerMutex.Lock()
	defer loggerMutex.Unlock()
	if lg, ok := loggerMap[name]; ok {
		return lg
	}
	lg := &Logger{
		name:   name,
		writer: &ConsoleWriter{},
	}

	loggerMap[name] = lg
	return lg
}

// SetLevel sets the log level
func SetLevel(level LogLevel) {
	logLevel = level
}

// GetLogLevel get global log level
func GetLogLevel() LogLevel {
	return logLevel
}

// GetLevel get global log level and return string
func GetLevel() string {
	return logLevel.String()
}

// SetFormat sets the log format
func SetFormat(format LogFormat) {
	logFormat = format
}

// GetLogFormat get global log format
func GetLogFormat() LogFormat {
	return logFormat
}

// GetFormat get global log format and return string
func GetFormat() string {
	return logFormat.String()
}

// Colored enable colored level string when use console writer
func Colored() {
	colored = true
}

// IsColored returns whether to enable colored when use console writer
func IsColored() bool {
	return colored
}

// StringToLevel turns string to LogLevel
func StringToLevel(level string) LogLevel {
	switch level {
	case "DEBUG":
		return DEBUG
	case "INFO":
		return INFO
	case "WARN":
		return WARN
	case "ERROR":
		return ERROR
	default:
		return DEBUG
	}
}

// SetCallerSkip sets the caller skip
func SetCallerSkip(skip int) {
	callerSkip = skip
}

// SetCallerFlag enable/disable caller string when write log
func SetCallerFlag(flag bool) {
	callerFlag = flag
}

// CallerFlag returns the caller's state string when writing to the log
func CallerFlag() bool {
	return callerFlag
}

// SetName sets the log name
func (l *Logger) SetName(name string) {
	l.name = name
}

// Name return the log name
func (l *Logger) Name() string {
	return l.name
}

// SetPrefix sets the log line prefix
func (l *Logger) SetPrefix(prefix string) {
	l.prefix = prefix
}

// Prefix returns the log line prefix
func (l *Logger) Prefix() string {
	return l.prefix
}

// SetFileRoller sets the file rolled by size in MB, with max num of files.
func (l *Logger) SetFileRoller(logpath string, num int, sizeMB int) error {
	if err := os.MkdirAll(logpath, 0755); err != nil {
		panic(err)
	}
	w := NewRollFileWriter(logpath, l.name, num, sizeMB)
	l.writer = w
	return nil
}

// IsConsoleWriter returns whether is consoleWriter or not.
func (l *Logger) IsConsoleWriter() bool {
	return reflect.TypeOf(l.writer) == reflect.TypeOf(&ConsoleWriter{})
}

// SetWriter sets the writer to the logger.
func (l *Logger) SetWriter(w LogWriter) {
	l.writer = w
}

// Writer return the log LogWriter.
func (l *Logger) Writer() LogWriter {
	return l.writer
}

// SetDayRoller sets the logger to rotate by day, with max num files.
func (l *Logger) SetDayRoller(logpath string, num int) error {
	if err := os.MkdirAll(logpath, 0755); err != nil {
		return err
	}
	w := NewDateWriter(logpath, l.name, DAY, num)
	l.writer = w
	return nil
}

// SetHourRoller sets the logger to rotate by hour, with max num files.
func (l *Logger) SetHourRoller(logpath string, num int) error {
	if err := os.MkdirAll(logpath, 0755); err != nil {
		return err
	}
	w := NewDateWriter(logpath, l.name, HOUR, num)
	l.writer = w
	return nil
}

// SetConsole sets the logger with console writer.
func (l *Logger) SetConsole() {
	l.writer = &ConsoleWriter{}
}

// Debug logs interface in debug loglevel.
func (l *Logger) Debug(v ...interface{}) {
	l.Writef(0, DEBUG, "", v)
}

// Info logs interface in Info loglevel.
func (l *Logger) Info(v ...interface{}) {
	l.Writef(0, INFO, "", v)
}

// Warn logs interface in warning loglevel
func (l *Logger) Warn(v ...interface{}) {
	l.Writef(0, WARN, "", v)
}

// Error logs interface in Error loglevel
func (l *Logger) Error(v ...interface{}) {
	l.Writef(0, ERROR, "", v)
}

// Trace log
func (l *Logger) Trace(msg string) {
	buf := bytes.NewBuffer(nil)
	if l.writer.NeedPrefix() {
		fmt.Fprintf(buf, "%s|", time.Now().Format("2006-01-02 15:04:05"))
	}
	fmt.Fprint(buf, msg)
	if l.writer.NeedPrefix() {
		buf.WriteByte('\n')
	}
	l.WriteLog(buf.Bytes())
}

// Debugf logs interface in debug loglevel with formating string
func (l *Logger) Debugf(format string, v ...interface{}) {
	l.Writef(0, DEBUG, format, v)
}

// Infof logs interface in Infof loglevel with formating string
func (l *Logger) Infof(format string, v ...interface{}) {
	l.Writef(0, INFO, format, v)
}

// Warnf logs interface in warning loglevel with formating string
func (l *Logger) Warnf(format string, v ...interface{}) {
	l.Writef(0, WARN, format, v)
}

// Errorf logs interface in Error loglevel with formating string
func (l *Logger) Errorf(format string, v ...interface{}) {
	l.Writef(0, ERROR, format, v)
}

func (l *Logger) Writef(depth int, level LogLevel, format string, v []interface{}) {
	if level < logLevel {
		return
	}

	if logFormat == Json {
		logQueue <- l.writeJson(depth, level, format, v)
	} else {
		logQueue <- l.writeLine(depth, level, format, v)
	}
}

func (l *Logger) writeLine(depth int, level LogLevel, format string, v []interface{}) *logValue {
	buf := bytes.NewBuffer(nil)
	if l.writer.NeedPrefix() {
		fmt.Fprintf(buf, "%s|", time.Now().Format("2006-01-02 15:04:05.000"))
		if len(l.prefix) > 0 {
			fmt.Fprintf(buf, "%s|", l.prefix)
		}

		if callerFlag {
			pc, file, line, ok := runtime.Caller(depth + callerSkip)
			if !ok {
				file = "???"
				line = 0
			} else {
				file = filepath.Base(file)
			}
			fmt.Fprintf(buf, "%s:%s:%d|", file, FuncName(runtime.FuncForPC(pc)), line)
		}
		if IsColored() && l.IsConsoleWriter() {
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
	if l.writer.NeedPrefix() {
		buf.WriteByte('\n')
	}
	return &logValue{value: buf.Bytes(), writer: l.writer}
}

func (l *Logger) writeJson(depth int, level LogLevel, format string, v []interface{}) *logValue {
	log := JsonLog{}
	log.Pre = l.prefix
	log.Time = time.Now().Format("2006-01-02 15:04:05.000")

	if callerFlag {
		pc, file, line, ok := runtime.Caller(depth + callerSkip)
		if !ok {
			file = "???"
			line = 0
		} else {
			file = filepath.Base(file)
		}
		log.Func = FuncName(runtime.FuncForPC(pc))
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
	return &logValue{value: buf.Bytes(), writer: l.writer}
}

// WriteLog write log into log files ignore the log level and log prefix
func (l *Logger) WriteLog(msg []byte) {
	logQueue <- &logValue{value: msg, writer: l.writer}
}

func FuncName(f *runtime.Func) string {
	name := f.Name()
	idx := strings.LastIndexByte(name, '/')
	if idx != -1 {
		name = name[idx:]
		idx = strings.IndexByte(name, '.')
		if idx != -1 {
			name = strings.TrimPrefix(name[idx:], ".")
		}
	}
	return name
}

// FlushLogger flush all log to disk.
func FlushLogger() {
	syncCancel()
	select {
	case <-time.After(waitFlushTimeout):
	case <-asyncDone.Done():
	}
}

func flushLog() {
	for {
		select {
		case v := <-logQueue:
			v.writer.Write(v.value)
		default:
			select {
			case v := <-logQueue:
				v.writer.Write(v.value)
			case <-syncDone.Done():
				asyncCancel()
				return
			}
		}
	}
}
