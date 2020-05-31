package tars

import (
	"os"

	"github.com/TarsCloud/TarsGo/tars/protocol/res/nodef"
)

// NodeFHelper is helper struct.
type NodeFHelper struct {
	comm *Communicator
	si   nodef.ServerInfo
	sf   *nodef.ServerF
}

// SetNodeInfo sets node information with communicator, node name, app name, server and container name
func (n *NodeFHelper) SetNodeInfo(comm *Communicator, node string, app string, server string, container string) {
	if node == "" {
		return
	}
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

// KeepAlive sends the keepalive pacakage to the node.
func (n *NodeFHelper) KeepAlive(adapter string) {
	if n.sf == nil {
		return
	}
	n.si.Pid = int32(os.Getpid())
	n.si.Adapter = adapter
	_, err := n.sf.KeepAlive(&n.si)
	if err != nil {
		TLOG.Error("keepalive fail:", adapter)
	}
}

// ReportVersion report the tars version to the node.
func (n *NodeFHelper) ReportVersion(version string) {
	if n.sf == nil {
		return
	}
	_, err := n.sf.ReportVersion(n.si.Application, n.si.ServerName, version)
	if err != nil {
		TLOG.Error("report Version fail:")
	}
}
