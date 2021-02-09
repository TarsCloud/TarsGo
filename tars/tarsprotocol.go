package tars

import (
	"bytes"
	"context"
	"encoding/binary"
	"time"

	"github.com/TarsCloud/TarsGo/tars/protocol"
	"github.com/TarsCloud/TarsGo/tars/protocol/codec"
	"github.com/TarsCloud/TarsGo/tars/protocol/res/basef"
	"github.com/TarsCloud/TarsGo/tars/protocol/res/requestf"
	"github.com/TarsCloud/TarsGo/tars/util/current"
)

type dispatch interface {
	Dispatch(context.Context, interface{}, *requestf.RequestPacket, *requestf.ResponsePacket, bool) error
}

// TarsProtocol is struct for dispatch with tars protocol.
type TarsProtocol struct {
	dispatcher  dispatch
	serverImp   interface{}
	withContext bool
}

// NewTarsProtocol return a TarsProtocol with dipatcher and implement interface.
// withContext explain using context or not.
func NewTarsProtocol(dispatcher dispatch, imp interface{}, withContext bool) *TarsProtocol {
	s := &TarsProtocol{dispatcher: dispatcher, serverImp: imp, withContext: withContext}
	return s
}

// Invoke puts the request as []byte and call the dispather, and then return the response as []byte.
func (s *TarsProtocol) Invoke(ctx context.Context, req []byte) (rsp []byte) {
	defer CheckPanic()
	reqPackage := requestf.RequestPacket{}
	rspPackage := requestf.ResponsePacket{}
	is := codec.NewReader(req[4:])
	reqPackage.ReadFrom(is)

	if reqPackage.HasMessageType(basef.TARSMESSAGETYPEDYED) {
		if dyeingKey, ok := reqPackage.Status[current.STATUS_DYED_KEY]; ok {
			if ok := current.SetDyeingKey(ctx, dyeingKey); !ok {
				TLOG.Error("dyeing-debug: set dyeing key in current status error, dyeing key:", dyeingKey)
			}
		}
	}

	if reqPackage.CPacketType == basef.TARSONEWAY {
		defer func() func() {
			beginTime := time.Now().UnixNano() / 1e6
			return func() {
				endTime := time.Now().UnixNano() / 1e6
				ReportStatFromServer(reqPackage.SFuncName, "one_way_client", rspPackage.IRet, endTime-beginTime)
			}
		}()()
	}

	if reqPackage.SFuncName == "tars_ping" {
		rspPackage.IVersion = reqPackage.IVersion
		//rspPackage.CPacketType = basef.TARSNORMAL
		rspPackage.IRequestId = reqPackage.IRequestId
		rspPackage.IRet = 0
	} else {
		var err error
		if s.withContext {
			ok := current.SetRequestStatus(ctx, reqPackage.Status)
			if !ok {
				TLOG.Error("Set reqeust status in context fail!")
			}
			ok = current.SetRequestContext(ctx, reqPackage.Context)
			if !ok {
				TLOG.Error("Set request context in context fail!")
			}
		}
		if allFilters.sf != nil {
			err = allFilters.sf(ctx, s.dispatcher.Dispatch, s.serverImp, &reqPackage, &rspPackage, s.withContext)
		} else if sf := getMiddlewareServerFilter(); sf != nil {
			err = sf(ctx, s.dispatcher.Dispatch, s.serverImp, &reqPackage, &rspPackage, s.withContext)
		} else {
			// execute pre server filters
			for i, v := range allFilters.preSfs {
				err = v(ctx, s.dispatcher.Dispatch, s.serverImp, &reqPackage, &rspPackage, s.withContext)
				if err != nil {
					TLOG.Errorf("Pre filter error, No.%v, err: %v", i, err.Error())
				}
			}
			err = s.dispatcher.Dispatch(ctx, s.serverImp, &reqPackage, &rspPackage, s.withContext)
			// execute post server filters
			for i, v := range allFilters.postSfs {
				err = v(ctx, s.dispatcher.Dispatch, s.serverImp, &reqPackage, &rspPackage, s.withContext)
				if err != nil {
					TLOG.Errorf("Post filter error, No.%v, err: %v", i, err.Error())
				}
			}
		}
		if err != nil {
			TLOG.Errorf("RequestID:%d, Found err: %v", reqPackage.IRequestId, err)
			//rspPackage.IVersion = basef.TARSVERSION
			rspPackage.IVersion = reqPackage.IVersion
			rspPackage.CPacketType = basef.TARSNORMAL
			rspPackage.IRequestId = reqPackage.IRequestId
			rspPackage.IRet = 1
			rspPackage.SResultDesc = err.Error()
			if tarsErr, ok := err.(*Error); ok {
				rspPackage.IRet = tarsErr.Code
			}
		}
	}

	//return ctype
	rspPackage.CPacketType = reqPackage.CPacketType
	ok := current.SetPacketTypeFromContext(ctx, rspPackage.CPacketType)
	if !ok {
		TLOG.Error("SetPacketType in context fail!")
	}

	return s.rsp2Byte(&rspPackage)
}

func (s *TarsProtocol) req2Byte(rsp *requestf.ResponsePacket) []byte {
	req := requestf.RequestPacket{}
	req.IVersion = rsp.IVersion
	req.IRequestId = rsp.IRequestId
	req.IMessageType = rsp.IMessageType
	req.CPacketType = rsp.CPacketType
	req.Context = rsp.Context
	req.Status = rsp.Status
	req.SBuffer = rsp.SBuffer

	os := codec.NewBuffer()
	req.WriteTo(os)
	bs := os.ToBytes()
	sbuf := bytes.NewBuffer(nil)
	sbuf.Write(make([]byte, 4))
	sbuf.Write(bs)
	len := sbuf.Len()
	binary.BigEndian.PutUint32(sbuf.Bytes(), uint32(len))
	return sbuf.Bytes()
}

func (s *TarsProtocol) rsp2Byte(rsp *requestf.ResponsePacket) []byte {
	if rsp.IVersion == basef.TUPVERSION {
		return s.req2Byte(rsp)
	}
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

// ParsePackage parse the []byte according to the tars protocol.
// returns header length and package integrity condition (PACKAGE_LESS | PACKAGE_FULL | PACKAGE_ERROR)
func (s *TarsProtocol) ParsePackage(buff []byte) (int, int) {
	return protocol.TarsRequest(buff)
}

// InvokeTimeout indicates how to deal with timeout.
func (s *TarsProtocol) InvokeTimeout(pkg []byte) []byte {
	rspPackage := requestf.ResponsePacket{}
	rspPackage.IRet = 1
	rspPackage.SResultDesc = "server invoke timeout"
	return s.rsp2Byte(&rspPackage)
}

// GetCloseMsg return a package to close connection
func (s *TarsProtocol) GetCloseMsg() []byte {
	rspPackage := requestf.ResponsePacket{}
	rspPackage.IRequestId = 0
	rspPackage.SResultDesc = reconnectMsg
	return s.rsp2Byte(&rspPackage)
}

// DoClose be called when close connection
func (s *TarsProtocol) DoClose(ctx context.Context) {
	TLOG.Debug("DoClose!")
}

