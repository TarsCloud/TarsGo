package tars

import (
	"encoding/base64"
	"fmt"
	"path/filepath"
	"strconv"
	"time"

	"github.com/TarsCloud/TarsGo/tars/util/rogger"
	"github.com/TarsCloud/TarsGo/tars/util/sync"
)

var (
	traceLogger *rogger.Logger
	loggerOnce  sync.Once
)

func Trace(traceKey, annotation, client, server, funcName string, ret int32, data, ex string) {
	loggerOnce.Do(func() (err error) {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("loggerOnce err: %+v", r)
			}
		}()
		traceLogger = GetRemoteLogger("_t_trace_")
		return nil
	})
	if traceLogger == nil {
		TLOG.Error("trace logger init error")
		return
	}
	msg := traceKey + "|" + annotation + "|" + client + "|" + server + "|" + funcName + "|" + strconv.FormatInt(time.Now().UnixNano()/1e6, 10) + "|" + strconv.Itoa(int(ret)) + "|" + base64.StdEncoding.EncodeToString([]byte(data)) + "|" + ex
	traceLogger.Trace(msg)
}

// GetLogger Get a logger
func GetLogger(name string) *rogger.Logger {
	logPath, cfg, lg := getLogger(name)
	// if the default writer is not ConsoleWriter, the writer has already been configured
	if !lg.IsConsoleWriter() {
		return lg
	}
	if cfg == nil {
		return lg
	}
	lg.SetFileRoller(logPath, int(cfg.LogNum), int(cfg.LogSize))
	return lg
}

func getLogger(name string) (logPath string, cfg *serverConfig, lg *rogger.Logger) {
	cfg = GetServerConfig()
	if cfg == nil {
		return "", nil, rogger.GetLogger(name)
	}
	if name == "" {
		name = cfg.App + "." + cfg.Server
	} else {
		name = cfg.App + "." + cfg.Server + "_" + name
	}
	logPath = filepath.Join(cfg.LogPath, cfg.App, cfg.Server)
	lg = rogger.GetLogger(name)
	return
}

// GetDayLogger Get a logger roll by day
func GetDayLogger(name string, numDay int) *rogger.Logger {
	logPath, _, lg := getLogger(name)
	// if the default writer is not ConsoleWriter, the writer has already been configured
	if !lg.IsConsoleWriter() {
		return lg
	}
	lg.SetDayRoller(logPath, numDay)
	return lg
}

// GetHourLogger Get a logger roll by hour
func GetHourLogger(name string, numHour int) *rogger.Logger {
	logPath, _, lg := getLogger(name)
	// if the default writer is not ConsoleWriter, the writer has already been configured
	if !lg.IsConsoleWriter() {
		return lg
	}
	lg.SetHourRoller(logPath, numHour)
	return lg
}

// GetRemoteLogger returns a remote logger
func GetRemoteLogger(name string) *rogger.Logger {
	cfg := GetServerConfig()
	lg := rogger.GetLogger(name)
	if !lg.IsConsoleWriter() {
		return lg
	}
	remoteWriter := NewRemoteTimeWriter()
	var set string
	if cfg.Enableset {
		set = cfg.Setdivision
	}

	remoteWriter.InitServerInfo(cfg.App, cfg.Server, name, set)
	lg.SetWriter(remoteWriter)
	return lg
}
