package tars

import (
	"testing"

	"github.com/TarsCloud/TarsGo/tars/protocol/res/configf"
)

func TestStringToProxy(t *testing.T) {
	comm := NewCommunicator()
	comm.SetProperty("locator", "tars.tarsregistry.QueryObj@tcp -h 172.0.0.1 -t 60000 -p 17890")
	prx := new(configf.Config)
	comm.StringToProxy("tars.tarsconfig.ConfigObj", prx, WithSet("test.test.1"))
}
