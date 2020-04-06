package tars

import (
	"time"

	"github.com/TarsCloud/TarsGo/tars/protocol/res/requestf"
)

// HashType is the hash type
type HashType int

// HashType enum
const (
	ModHash HashType = iota
	ConsistentHash
)

// Message is a struct contains servant information
type Message struct {
	Req  *requestf.RequestPacket
	Resp *requestf.ResponsePacket

	Ser *ServantProxy
	Adp *AdapterProxy

	BeginTime int64
	EndTime   int64
	Status    int32

	hashCode uint32
	hashType HashType
	isHash   bool
}

// Init define the begintime
func (m *Message) Init() {
	m.BeginTime = time.Now().UnixNano() / 1e6
}

// End define the endtime
func (m *Message) End() {
	m.EndTime = time.Now().UnixNano() / 1e6
}

// Cost calculate the cost time
func (m *Message) Cost() int64 {
	return m.EndTime - m.BeginTime
}

// SetHash set hash code
func (m *Message) SetHash(code uint32, h HashType) {
	m.hashCode = code
	m.hashType = h
	m.isHash = true
}
