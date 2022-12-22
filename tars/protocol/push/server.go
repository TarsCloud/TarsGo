package push

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"net"

	"github.com/TarsCloud/TarsGo/tars"
	"github.com/TarsCloud/TarsGo/tars/protocol/codec"
	"github.com/TarsCloud/TarsGo/tars/protocol/res/requestf"
	"github.com/TarsCloud/TarsGo/tars/transport"
	"github.com/TarsCloud/TarsGo/tars/util/current"
	"github.com/TarsCloud/TarsGo/tars/util/tools"
)

// PushServer defines the pushing server
type PushServer interface {
	OnConnect(ctx context.Context, req []byte) []byte
	OnClose(ctx context.Context)
}

type serverProtocol struct {
	tars.Protocol
	s PushServer
}

// Send push message to client
func Send(ctx context.Context, data []byte) error {
	conn, udpAddr, ok := current.GetRawConn(ctx)
	if !ok {
		return fmt.Errorf("connection not found")
	}
	rsp := &requestf.ResponsePacket{
		SBuffer: tools.ByteToInt8(data),
	}
	rspData := response2Bytes(rsp)
	var err error
	if udpAddr != nil {
		udpConn, _ := conn.(*net.UDPConn)
		_, err = udpConn.WriteToUDP(rspData, udpAddr)
	} else {
		_, err = conn.Write(rspData)
	}
	return err
}

// NewServer return a server for pushing message
func NewServer(s PushServer) transport.ServerProtocol {
	return &serverProtocol{Protocol: tars.Protocol{}, s: s}
}

func (s *serverProtocol) DoClose(ctx context.Context) {
	s.s.OnClose(ctx)
}

// Invoke process request and send response
func (s *serverProtocol) Invoke(ctx context.Context, reqBytes []byte) []byte {
	req := &requestf.RequestPacket{}
	rsp := &requestf.ResponsePacket{}
	is := codec.NewReader(reqBytes[4:])
	if err := req.ReadFrom(is); err != nil {
		rsp.IRet = 1
		rsp.SResultDesc = "decode request package error"
	} else {
		rsp.IVersion = req.IVersion
		rsp.CPacketType = req.CPacketType
		rsp.IRequestId = req.IRequestId
		if req.SFuncName != "tars_ping" {
			rspData := s.s.OnConnect(ctx, tools.Int8ToByte(req.SBuffer))
			rsp.SBuffer = tools.ByteToInt8(rspData)
		}
	}
	return response2Bytes(rsp)
}

func response2Bytes(rsp *requestf.ResponsePacket) []byte {
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
