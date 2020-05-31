package rogger

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"time"
)

//DEBUG loglevel
const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	OFF
)

var (
	logLevel = DEBUG
	colored  = false

	logQueue    = make(chan *logValue, 10000)
	loggerMutex sync.Mutex
	loggerMap   = make(map[string]*Logger)
	writeDone   = make(chan bool)
	callerSkip  = 2
	callerFlag  = true

	waitFlushTimeout       = time.Second
	syncDone, syncCancel   = context.WithCancel(context.Background())
	asyncDone, asyncCancel = context.WithCancel(context.Background())
)

// Logger is the struct with name and wirter.
type Logger struct {
	name   string
	writer LogWriter
}

// LogLevel is uint8 type
type LogLevel uint8

type logValue struct {
	level  LogLevel
	value  []byte
	fileNo string
	writer LogWriter
}

func init() {
	go flushLog()
}

// String turns the LogLevel to string.
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

// Colored enable colored level string when use console writer
func (lv *LogLevel) coloredString() string {
	switch *lv {
	case DEBUG:
		return "\x1b[34mDEBUG\x1b[0m" //blue
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

// Colored enable colored level string when use console writer
func Colored() {
	colored = true
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

// SetLogName sets the log name
func (l *Logger) SetLogName(name string) {
	l.name = name
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
	if reflect.TypeOf(l.writer) == reflect.TypeOf(&ConsoleWriter{}) {
		return true
	}
	return false
}

// SetWriter sets the writer to the logger.
func (l *Logger) SetWriter(w LogWriter) {
	l.writer = w
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

	buf := bytes.NewBuffer(nil)
	if l.writer.NeedPrefix() {
		fmt.Fprintf(buf, "%s|", time.Now().Format("2006-01-02 15:04:05.000"))

		if callerFlag {
			pc, file, line, ok := runtime.Caller(depth + callerSkip)
			if !ok {
				file = "???"
				line = 0
			} else {
				file = filepath.Base(file)
			}
			fmt.Fprintf(buf, "%s:%s:%d|", file, getFuncName(runtime.FuncForPC(pc).Name()), line)
		}
		if colored && l.IsConsoleWriter() {
			buf.WriteString(level.coloredString())
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
	logQueue <- &logValue{value: buf.Bytes(), writer: l.writer}
}

func getFuncName(name string) string {
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

// WriteLog write log into log files ignore the log level and log prefix
func (l *Logger) WriteLog(msg []byte) {
	logQueue <- &logValue{value: msg, writer: l.writer}
}

// FlushLogger flushs all log to disk.
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
