package tars

import (
	"context"
	"time"

	"github.com/TarsCloud/TarsGo/tars/model"
	"github.com/TarsCloud/TarsGo/tars/protocol/res/basef"
	"github.com/TarsCloud/TarsGo/tars/protocol/res/requestf"
	"github.com/TarsCloud/TarsGo/tars/selector"
	"github.com/TarsCloud/TarsGo/tars/util/current"
	"github.com/TarsCloud/TarsGo/tars/util/tools"
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
	Async    bool
	Callback model.Callback
	RespCh   chan *requestf.ResponsePacket
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

func newMessage(ctx context.Context, cType byte,
	sFuncName string,
	buf []byte,
	status map[string]string,
	reqContext map[string]string,
	resp *requestf.ResponsePacket,
	s *ServantProxy) *Message {

	// 将ctx中的dyeing信息传入到request中
	var msgType int32
	if dyeingKey, ok := current.GetDyeingKey(ctx); ok {
		TLOG.Debug("dyeing debug: find dyeing key:", dyeingKey)
		if status == nil {
			status = make(map[string]string)
		}
		status[current.StatusDyedKey] = dyeingKey
		msgType |= basef.TARSMESSAGETYPEDYED
	}

	// 将ctx中的trace信息传入到request中
	if trace, ok := current.GetTarsTrace(ctx); ok && trace.Call() {
		traceKey := trace.GetTraceFullKey(false)
		TLOG.Debug("trace debug: find trace key:", traceKey)
		if status == nil {
			status = make(map[string]string)
		}
		status[current.StatusTraceKey] = traceKey
		msgType |= basef.TARSMESSAGETYPETRACE
	}

	req := requestf.RequestPacket{
		IVersion:     s.version,
		CPacketType:  int8(cType),
		IMessageType: msgType,
		IRequestId:   s.genRequestID(),
		SServantName: s.name,
		SFuncName:    sFuncName,
		ITimeout:     int32(s.syncTimeout),
		SBuffer:      tools.ByteToInt8(buf),
		Context:      reqContext,
		Status:       status,
	}
	msg := &Message{Req: &req, Ser: s, Resp: resp}
	msg.Init()

	if ok, hashType, hashCode, isHash := current.GetClientHash(ctx); ok {
		msg.isHash = isHash
		msg.hashType = HashType(hashType)
		msg.hashCode = hashCode
	}

	return msg
}
