package log

import (
	"path/filepath"

	"github.com/TarsCloud/TarsGo/tars"
)

// GetLogger Get a logger
func GetLogger(name string) *Logger {
	logPath, lg := getLogger(name)
	// if the default writer is not ConsoleWriter, the writer has already been configured
	if !lg.IsConsoleWriter() {
		return lg
	}
	if logPath == "" {
		return lg
	}
	cfg := tars.GetServerConfig()
	lg.SetFileRoller(logPath, int(cfg.LogNum), int(cfg.LogSize))
	return lg
}

func getLogger(name string) (logPath string, lg *Logger) {
	cfg := tars.GetServerConfig()
	if cfg == nil {
		return "", GetCtxLogger(name)
	}
	if name == "" {
		name = cfg.App + "." + cfg.Server
	} else {
		name = cfg.App + "." + cfg.Server + "_" + name
	}
	logPath = filepath.Join(cfg.LogPath, cfg.App, cfg.Server)
	lg = GetCtxLogger(name)
	return logPath, lg
}

// GetDayLogger Get a logger roll by day
func GetDayLogger(name string, numDay int) *Logger {
	logPath, lg := getLogger(name)
	// if the default writer is not ConsoleWriter, the writer has already been configured
	if !lg.IsConsoleWriter() {
		return lg
	}
	lg.SetDayRoller(logPath, numDay)
	return lg
}

// GetHourLogger Get a logger roll by hour
func GetHourLogger(name string, numHour int) *Logger {
	logPath, lg := getLogger(name)
	// if the default writer is not ConsoleWriter, the writer has already been configured
	if !lg.IsConsoleWriter() {
		return lg
	}
	lg.SetHourRoller(logPath, numHour)
	return lg
}

// GetRemoteLogger returns a remote logger
func GetRemoteLogger(name string) *Logger {
	cfg := tars.GetServerConfig()
	lg := GetCtxLogger(name)
	if !lg.IsConsoleWriter() {
		return lg
	}
	remoteWriter := tars.NewRemoteTimeWriter()
	var set string
	if cfg.Enableset {
		set = cfg.Setdivision
	}

	remoteWriter.InitServerInfo(cfg.App, cfg.Server, name, set)
	lg.SetWriter(remoteWriter)
	return lg
}
