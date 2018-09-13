package tars

import (
	"github.com/TarsCloud/TarsGo/tars/protocol/res/basef"
	"github.com/TarsCloud/TarsGo/tars/protocol/res/requestf"
	"time"
)

type Message struct {
	Req  *requestf.RequestPacket
	Resp *requestf.ResponsePacket

	Obj *ObjectProxy
	Ser *ServantProxy
	Adp *AdapterProxy

	BeginTime int64
	EndTime   int64
	Status    int

	hashCode int64
	isHash   bool
}

func (m *Message) Init() {
	m.BeginTime = time.Now().UnixNano() / 1000000
}

func (m *Message) End() {
	m.Status = int(basef.TARSSERVERSUCCESS)
	m.EndTime = time.Now().UnixNano() / 1000000
}

func (m *Message) Cost() int64 {
	return m.EndTime - m.BeginTime
}

func (m *Message) SetHashCode(code int64) {
	m.hashCode = code
	m.isHash = true
}
