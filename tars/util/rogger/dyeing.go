package rogger

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/TarsCloud/TarsGo/tars/util/current"
)

const (
	Admin = "dyeUserManage"
)

var (
	mutex          sync.RWMutex
	dyeingUserMap  = make(map[string]bool)
	dyeingLogQueue = make(chan *dyeingLogValue, 10000)
	dyeingErrorLog = GetLogger("TLOG")
)

type dyeingLogValue struct {
	Level     LogLevel
	Value     []byte
	DyeingKey string
	ExtInfo   interface{}
}

// GetDyeingLogQue get dyeingLogQueue, which will contain dyeing log if dyeing switch is on(by call DyeingSwitch).
// If dyeing switch is on, user must guarantee the dyeingLogQueue can be continuous consumed in case of goroutine blocked.
func GetDyeingLogQueue() *chan *dyeingLogValue {
	return &dyeingLogQueue
}

func (l *Logger) DyeingDebug(ctx context.Context, ext interface{}, v ...interface{}) {
	l.DyeingWritef(ctx, 0, DEBUG, ext, "", v)
}

func (l *Logger) DyeingInfo(ctx context.Context, ext interface{}, v ...interface{}) {
	l.DyeingWritef(ctx, 0, INFO, ext, "", v)
}

func (l *Logger) DyeingWarn(ctx context.Context, ext interface{}, v ...interface{}) {
	l.DyeingWritef(ctx, 0, WARN, ext, "", v)
}

func (l *Logger) DyeingError(ctx context.Context, ext interface{}, v ...interface{}) {
	l.DyeingWritef(ctx, 0, ERROR, ext, "", v)
}

func (l *Logger) DyeingDebugf(ctx context.Context, ext interface{}, format string, v ...interface{}) {
	l.DyeingWritef(ctx, 0, DEBUG, ext, format, v)
}

func (l *Logger) DyeingInfof(ctx context.Context, ext interface{}, format string, v ...interface{}) {
	l.DyeingWritef(ctx, 0, INFO, ext, format, v)
}

func (l *Logger) DyeingWarnf(ctx context.Context, ext interface{}, format string, v ...interface{}) {
	l.DyeingWritef(ctx, 0, WARN, ext, format, v)
}

func (l *Logger) DyeingErrorf(ctx context.Context, ext interface{}, format string, v ...interface{}) {
	l.DyeingWritef(ctx, 0, ERROR, ext, format, v)
}

func (l *Logger) DyeingWritef(ctx context.Context, depth int, level LogLevel, ext interface{}, format string, v []interface{}) {
	l.Writef(depth+1, level, format, v)

	dyeingKey, ok := current.GetDyeingKey(ctx)
	if !ok {
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

	select {
	case dyeingLogQueue <- &dyeingLogValue{Value: buf.Bytes(), DyeingKey: dyeingKey, ExtInfo: ext}:
	default:
		dyeingErrorLog.Error("dyeingLogQueue is full")
	}
}

// AddDyeingUser add dyeing key to dyeingUserMap. key is separated by ','
func AddDyeingUser(key string) error {
	v := strings.Split(key, ",")
	if len(v) == 0 {
		return errors.New("add dyeingkey, but no dyeingkey found")
	}

	mutex.Lock()
	defer mutex.Unlock()
	for index := range v {
		dyeingUserMap[v[index]] = true
	}
	return nil
}

// RemoveDyeingUser remove dyeing key from dyeingUserMap. key is separated by ','
func RemoveDyeingUser(key string) error {
	mutex.Lock()
	defer mutex.Unlock()

	v := strings.Split(key, ",")
	for i := range v {
		delete(dyeingUserMap, v[i])
	}

	return nil
}

// SetDyeingUser set dyeing key to dyeingUserMap(overwrite)
func SetDyeingUser(v []string) {
	mutex.Lock()
	defer mutex.Unlock()

	dyeingUserMap = make(map[string]bool)
	for i := range v {
		dyeingUserMap[v[i]] = true
	}
}

// GetAllDyeingUser get all dyeing key from dyeingUserMap. key is separated by ','
func GetAllDyeingUser() []string {
	mutex.RLock()
	defer mutex.RUnlock()

	var arr []string
	for k, _ := range dyeingUserMap {
		arr = append(arr, k)
	}

	return arr
}

// IsDyeingUser return whether dyeingKey exist in dyeingUserMap
func IsDyeingUser(key string) bool {
	mutex.RLock()
	defer mutex.RUnlock()

	_, ok := dyeingUserMap[key]
	return ok
}

// HandleDyeingAdmin handle the dyeing key operation from admin.
// For example add a dyeing key. Send "dyeUserManage help" from admin, will see help.
func HandleDyeingAdmin(command string) (string, error) {
	var msg string
	cmd := strings.Split(command, " ")

	if len(cmd) <= 1 {
		msg = "dyeing param error see help"
		return msg, errors.New(msg)
	}

	if cmd[0] != "dyeUserManage" {
		msg = "taf frame err"
		return msg, errors.New(msg)
	}

	if cmd[1] == "help" {
		msg = "command format : dyeUserManage sub command [parameter]\n"
		msg += "sub command and parameter:\n"
		msg += "add dyeinguser                 : add single dyeing key\n"
		msg += "adds dyeinguser1,dyeinguser2...      : add multi dyeing keys\n"
		msg += "remove dyeinguser              : delete single dyeing key\n"
		msg += "removes dyeinguser1,dyeinguser2...   : delete multi dyeing keys\n"
		msg += "removeall                : delete all dyeing key\n"
		msg += "getall                   : get all dyeing key\n"
		return msg, nil
	}

	switch cmd[1] {
	case "add":
		if len(cmd) != 3 {
			msg = "dyeing param err see help"
			return msg, errors.New(msg)
		}

		err := AddDyeingUser(cmd[2])
		if err != nil {
			return err.Error(), err
		}
	case "adds":
		if len(cmd) != 3 {
			msg = "dyeing param err see help"
			return msg, errors.New(msg)
		}

		err := AddDyeingUser(cmd[2])
		if err != nil {
			return err.Error(), err
		}
	case "remove":
		if len(cmd) != 3 {
			msg = "dyeing param err see help"
			return msg, errors.New(msg)
		}
		err := RemoveDyeingUser(cmd[2])
		if err != nil {
			return err.Error(), err
		}
	case "removes":
		err := RemoveDyeingUser(cmd[2])
		if err != nil {
			return err.Error(), err
		}
	case "removeall":
		SetDyeingUser([]string{})
	case "getall":
		v := GetAllDyeingUser()
		msg = strings.Join(v, ",")
		return msg, nil
	default:
		msg = "param error see help"
		return msg, errors.New(msg)
	}

	return "OK", nil
}
