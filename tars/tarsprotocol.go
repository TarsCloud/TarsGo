package tars

import (
	"bytes"
	"encoding/binary"
	"github.com/TarsCloud/TarsGo/tars/protocol/codec"
	"github.com/TarsCloud/TarsGo/tars/protocol/res/basef"
	"github.com/TarsCloud/TarsGo/tars/protocol/res/requestf"
	"time"
)

type dispatch interface {
	Dispatch(interface{}, *requestf.RequestPacket, *requestf.ResponsePacket) error
}

type TarsProtocol struct {
	dispatcher dispatch
	serverImp  interface{}
}

func NewTarsProtocol(dispatcher dispatch, imp interface{}) *TarsProtocol {
	s := &TarsProtocol{dispatcher: dispatcher, serverImp: imp}
	return s
}

func (s *TarsProtocol) Invoke(req []byte) (rsp []byte) {
	defer checkPanic()
	reqPackage := requestf.RequestPacket{}
	rspPackage := requestf.ResponsePacket{}
	is := codec.NewReader(req)
	reqPackage.ReadFrom(is)
	TLOG.Debug("invoke:", reqPackage.IRequestId)
	if reqPackage.CPacketType == basef.TARSONEWAY {
		defer func() func() {
			beginTime := time.Now().UnixNano() / 1000000
			return func() {
				endTime := time.Now().UnixNano() / 1000000
				ReportStatFromServer(reqPackage.SFuncName, "one_way_client", rspPackage.IRet, endTime-beginTime)
			}
		}()()
	}
	err := s.dispatcher.Dispatch(s.serverImp, &reqPackage, &rspPackage)
	if err != nil {
		rspPackage.IRet = 1
		rspPackage.SResultDesc = err.Error()
	}
	return s.rsp2Byte(&rspPackage)
}

func (s *TarsProtocol) rsp2Byte(rsp *requestf.ResponsePacket) []byte {
	os := codec.NewBuffer()
	rsp.WriteTo(os)
	bs := os.ToBytes()
	sbuf := bytes.NewBuffer(nil)
	sbuf.Write(make([]byte, 4))
	sbuf.Write(bs)
	len := sbuf.Len()
	binary.BigEndian.PutUint32(sbuf.Bytes(), uint32(len))
	return sbuf.Bytes()
}

func (s *TarsProtocol) ParsePackage(buff []byte) (int, int) {
	return TarsRequest(buff)
}

func (s *TarsProtocol) InvokeTimeout(pkg []byte) []byte {
	rspPackage := requestf.ResponsePacket{}
	rspPackage.IRet = 1
	rspPackage.SResultDesc = "server invoke timeout"
	return s.rsp2Byte(&rspPackage)
}
