package tars

import "github.com/TarsCloud/TarsGo/tars/protocol/res/notifyf"

const (
	NotifyNormal = 0
	NotifyWarn   = 1
	NotifyError  = 2
)

// NotifyHelper is the helper struct for the Notify service.
type NotifyHelper struct {
	comm *Communicator
	tn   *notifyf.Notify
	tm   notifyf.ReportInfo
}

// SetNotifyInfo sets the communicator's notify info with communicator, notify name, app name, server name, and container name
func (n *NotifyHelper) SetNotifyInfo(comm *Communicator, notify string, app string, server string, container string) {
	n.comm = comm
	n.tn = new(notifyf.Notify)
	comm.StringToProxy(notify, n.tn)
	var set string
	if v, ok := comm.GetProperty("setdivision"); ok {
		set = v
	}
	n.tm = notifyf.ReportInfo{
		EType:      0,
		SApp:       app,
		SSet:       set,
		SContainer: container,
		SServer:    server,
		SMessage:   "",
		SThreadId:  "",
		ELevel:     0,
	}
}

// ReportNotifyInfo reports notify information with level and info
func (n *NotifyHelper) ReportNotifyInfo(level int32, info string) {
	n.tm.ELevel = notifyf.NOTIFYLEVEL(level)
	n.tm.SMessage = info
	TLOG.Debug(n.tm)
	if err := n.tn.ReportNotifyInfo(&n.tm, n.comm.Client.Context()); err != nil {
		TLOG.Errorf("ReportNotifyInfo err: %v", err)
	}
}

// ReportNotifyInfo reports notify information with level and info
func ReportNotifyInfo(level int32, info string) {
	svrCfg := GetServerConfig()
	if svrCfg.Notify == "" {
		return
	}
	comm := GetCommunicator()
	ha := new(NotifyHelper)
	ha.SetNotifyInfo(comm, svrCfg.Notify, svrCfg.App, svrCfg.Server, svrCfg.Container)
	defer func() {
		if err := recover(); err != nil {
			TLOG.Debug(err)
		}
	}()
	ha.ReportNotifyInfo(level, info)
}
