package tars

import (
	"github.com/TarsCloud/TarsGo/tars/protocol/res/nodef"
	"os"
)

// NodeFHelper struct
type NodeFHelper struct {
	comm *Communicator
	si   nodef.ServerInfo
	sf   *nodef.ServerF
}

// SetNodeInfo sets node info
// with Communicator comm, string node name, string app name, string server
func (n *NodeFHelper) SetNodeInfo(comm *Communicator, node string, app string, server string) {
	n.comm = comm
	n.sf = new(nodef.ServerF)
	comm.StringToProxy(node, n.sf)
	n.si = nodef.ServerInfo{
		app,
		server,
		int32(os.Getpid()),
		"",
		//"tars",
		//container,
	}
}

// KeepAlive keeps the NodeFHelper's ServerF alive
func (n *NodeFHelper) KeepAlive(adapter string) {
	n.si.Adapter = adapter
	_, err := n.sf.KeepAlive(&n.si)
	if err != nil {
		TLOG.Error("keepalive fail:", adapter)
	}
}

// ReportVersion reports the version with string verion
func (n *NodeFHelper) ReportVersion(version string) {
	_, err := n.sf.ReportVersion(n.si.Application, n.si.ServerName, version)
	if err != nil {
		TLOG.Error("report Version fail:")
	}
}
