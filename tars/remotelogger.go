package tars

import (
	"time"

	"github.com/TarsCloud/TarsGo/tars/protocol/res/logf"
)

// RemoteTimeWriter writer for writing remote log.
type RemoteTimeWriter struct {
	logInfo          *logf.LogInfo
	logs             chan string
	logPtr           *logf.Log
	reportSuccessPtr *PropertyReport
	reportFailPtr    *PropertyReport
	hasPrefix        bool
	comm             *Communicator
}

// NewRemoteTimeWriter new and init RemoteTimeWriter
func NewRemoteTimeWriter() *RemoteTimeWriter {
	rw := &RemoteTimeWriter{
		logInfo: &logf.LogInfo{
			SFormat:    "%Y%m%d",
			SConcatStr: "_",
		},
		logs:   make(chan string, remoteLogQueueSize),
		logPtr: new(logf.Log),
	}
	rw.EnableSuffix(true)
	rw.EnablePrefix(true)
	rw.SetSeparator("|")
	rw.SetPrefix(true)

	log := GetServerConfig().Log
	comm := GetCommunicator()
	comm.StringToProxy(log, rw.logPtr)
	rw.comm = comm
	go rw.Sync2remote()
	return rw
}

// Sync2remote syncs the log buffer to remote.
func (rw *RemoteTimeWriter) Sync2remote() {
	maxLen := remoteLogMaxNumOneTime
	interval := remoteLogInterval
	v := make([]string, 0, maxLen)
	for {
		select {
		case log := <-rw.logs:
			v = append(v, log)
			if len(v) >= maxLen {
				err := rw.sync2remote(v)
				if err != nil {
					TLOG.Error("sync to remote error")
					rw.reportFailPtr.Report(len(v))
				}
				rw.reportSuccessPtr.Report(len(v))
				v = make([]string, 0, maxLen) //reset the slice after syncing log to remote
			}
		case <-time.After(interval):
			if len(v) > 0 {
				err := rw.sync2remote(v)
				if err != nil {
					TLOG.Error("sync to remote error")
					rw.reportFailPtr.Report(len(v))
				}
				rw.reportSuccessPtr.Report(len(v))
				v = make([]string, 0, maxLen) //reset the slice after syncing log to remote
			}
		}
	}
}

func (rw *RemoteTimeWriter) sync2remote(s []string) error {
	err := rw.logPtr.LoggerbyInfo(rw.logInfo, s, rw.comm.Client.Context())
	return err
}

// InitServerInfo init the remote log server info.
func (rw *RemoteTimeWriter) InitServerInfo(app string, server string, filename string, setdivision string) {
	rw.logInfo.Appname = app
	rw.logInfo.Servername = server
	rw.logInfo.SFilename = filename
	rw.logInfo.Setdivision = setdivision

	serverInfo := app + "." + server + "." + filename
	failServerInfo := serverInfo + "_log_send_fail"
	failSum := NewSum()
	rw.reportFailPtr = CreatePropertyReport(failServerInfo, failSum)
	successServerInfo := serverInfo + "_log_send_succ"
	successSum := NewSum()
	rw.reportSuccessPtr = CreatePropertyReport(successServerInfo, successSum)
}

// EnableSuffix puts suffix after logs.
func (rw *RemoteTimeWriter) EnableSuffix(hasSuffix bool) {
	rw.logInfo.BHasSufix = hasSuffix
}

// EnablePrefix puts prefix before logs.
func (rw *RemoteTimeWriter) EnablePrefix(hasAppNamePrefix bool) {
	rw.logInfo.BHasAppNamePrefix = hasAppNamePrefix
}

// SetFileNameConcatStr sets the filename concat string.
func (rw *RemoteTimeWriter) SetFileNameConcatStr(s string) {
	rw.logInfo.SConcatStr = s

}

// SetSeparator set separator between logs.
func (rw *RemoteTimeWriter) SetSeparator(s string) {
	rw.logInfo.SSepar = s
}

// EnableSquareWrapper enables SquareBracket wrapper for the logs.
func (rw *RemoteTimeWriter) EnableSquareWrapper(hasSquareBracket bool) {
	rw.logInfo.BHasSquareBracket = hasSquareBracket
}

// SetLogType sets the log type.
func (rw *RemoteTimeWriter) SetLogType(logType string) {
	rw.logInfo.SLogType = logType

}

// InitFormat sets the log format.
func (rw *RemoteTimeWriter) InitFormat(s string) {
	rw.logInfo.SFormat = s
}

// NeedPrefix return if you need prefix for the logger.
func (rw *RemoteTimeWriter) NeedPrefix() bool {
	return rw.hasPrefix
}

// SetPrefix set if you need prefix for the logger.
func (rw *RemoteTimeWriter) SetPrefix(enable bool) {
	rw.hasPrefix = enable
}

// Write the logs to the buffer.
func (rw *RemoteTimeWriter) Write(b []byte) {
	s := string(b[:])
	select {
	case rw.logs <- s:
	default:
		TLOG.Error("remote log chan is full")
	}
}
