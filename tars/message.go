package tars

import (
	"time"

	"github.com/TarsCloud/TarsGo/tars/protocol/res/requestf"
	"github.com/TarsCloud/TarsGo/tars/selector"
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

// Init define the beginTime
func (m *Message) Init() {
	m.BeginTime = time.Now().UnixNano() / 1e6
}

// End define the endTime
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

func (m *Message) HashCode() uint32 {
	return m.hashCode
}

func (m *Message) HashType() selector.HashType {
	return selector.HashType(m.hashType)
}

func (m *Message) IsHash() bool {
	return m.isHash
}
