package tars

import (
	"path/filepath"

	"github.com/TarsCloud/TarsGo/tars/util/rogger"
)

// GetLogger Get a logger
func GetLogger(name string) *rogger.Logger {
	logPath, cfg, lg := getLogger(name)
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
	lg.SetDayRoller(logPath, numDay)
	return lg
}

// GetHourLogger Get a logger roll by hour
func GetHourLogger(name string, numHour int) *rogger.Logger {
	logPath, _, lg := getLogger(name)
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
