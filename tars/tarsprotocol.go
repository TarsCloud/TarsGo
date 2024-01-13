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

// Protocol is struct for dispatch with tars protocol.
type Protocol struct {
	app         *application
	dispatcher  dispatch
	serverImp   interface{}
	withContext bool
}

const (
	reconnectMsg = "_reconnect_"
)

// NewTarsProtocol return a TarsProtocol with dispatcher and implement interface.
// withContext explain using context or not.
func NewTarsProtocol(dispatcher dispatch, imp interface{}, withContext bool) *Protocol {
	s := &Protocol{dispatcher: dispatcher, serverImp: imp, withContext: withContext}
	return s
}

// Invoke puts the request as []byte and call the dispatcher, and then return the response as []byte.
func (s *Protocol) Invoke(ctx context.Context, req []byte) (rsp []byte) {
	defer CheckPanic()
	reqPackage := requestf.RequestPacket{}
	rspPackage := requestf.ResponsePacket{}
	is := codec.NewReader(req[4:])
	reqPackage.ReadFrom(is)

	recvPkgTs, ok := current.GetRecvPkgTsFromContext(ctx)
	if !ok {
		recvPkgTs = time.Now().UnixNano() / 1e6
	}

	// timeout delivery
	now := time.Now().UnixNano() / 1e6
	if reqPackage.ITimeout > 0 {
		sub := now - recvPkgTs // coroutine scheduling time difference
		timeout := int64(reqPackage.ITimeout) - sub
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(timeout)*time.Millisecond)
		defer cancel()
	}

	if reqPackage.HasMessageType(basef.TARSMESSAGETYPEDYED) {
		if dyeingKey, ok := reqPackage.Status[current.StatusDyedKey]; ok {
			if ok = current.SetDyeingKey(ctx, dyeingKey); !ok {
				TLOG.Error("dyeing-debug: set dyeing key in current status error, dyeing key:", dyeingKey)
			}
		}
	}

	// 处理TARS下的调用链追踪
	if reqPackage.HasMessageType(basef.TARSMESSAGETYPETRACE) {
		if traceKey, ok := reqPackage.Status[current.StatusTraceKey]; ok {
			TLOG.Info("[TARS] servant got a trace request, trace key:", traceKey)
			if ok = current.InitTarsTrace(ctx, traceKey); !ok {
				TLOG.Error("trace-debug: set trace key in current status error, trace key:", traceKey)
			}
		}
	}

	if reqPackage.CPacketType == basef.TARSONEWAY {
		defer func() {
			endTime := time.Now().UnixNano() / 1e6
			ReportStatFromServer(reqPackage.SFuncName, "one_way_client", rspPackage.IRet, endTime-recvPkgTs)
		}()
	} else if reqPackage.CPacketType == basef.TARSNORMAL {
		defer func() {
			endTime := time.Now().UnixNano() / 1e6
			ReportStatFromServer(reqPackage.SFuncName, "stat_from_server", rspPackage.IRet, endTime-recvPkgTs)
		}()
	}
	// timeout or tars_ping or error
	rspPackage.IVersion = reqPackage.IVersion
	rspPackage.IRequestId = reqPackage.IRequestId

	select {
	case <-ctx.Done():
		rspPackage.IRet = basef.TARSSERVERQUEUETIMEOUT
		rspPackage.SResultDesc = "server invoke timeout"
		ip, _ := current.GetClientIPFromContext(ctx)
		port, _ := current.GetClientPortFromContext(ctx)
		TLOG.Errorf("handle queue timeout, obj:%s, func:%s, recv time:%d, now:%d, timeout:%d, cost:%d,  addr:(%s:%s), reqId:%d, err: %v",
			reqPackage.SServantName, reqPackage.SFuncName, recvPkgTs, now, reqPackage.ITimeout, now-recvPkgTs, ip, port, reqPackage.IRequestId, ctx.Err())
	default:
		if reqPackage.SFuncName != "tars_ping" { // not tars_ping, normal business call branch
			if s.withContext {
				if ok = current.SetRequestStatus(ctx, reqPackage.Status); !ok {
					TLOG.Error("Set request status in context fail!")
				}
				if ok = current.SetRequestContext(ctx, reqPackage.Context); !ok {
					TLOG.Error("Set request context in context fail!")
				}
			}
			var err error
			if s.app.allFilters.sf != nil {
				err = s.app.allFilters.sf(ctx, s.dispatcher.Dispatch, s.serverImp, &reqPackage, &rspPackage, s.withContext)
			} else if sf := s.app.getMiddlewareServerFilter(); sf != nil {
				err = sf(ctx, s.dispatcher.Dispatch, s.serverImp, &reqPackage, &rspPackage, s.withContext)
			} else {
				// execute pre server filters
				for i, v := range s.app.allFilters.preSfs {
					err = v(ctx, s.dispatcher.Dispatch, s.serverImp, &reqPackage, &rspPackage, s.withContext)
					if err != nil {
						TLOG.Errorf("Pre filter error, No.%v, err: %v", i, err)
					}
				}
				// execute business server
				err = s.dispatcher.Dispatch(ctx, s.serverImp, &reqPackage, &rspPackage, s.withContext)
				// execute post server filters
				for i, v := range s.app.allFilters.postSfs {
					err = v(ctx, s.dispatcher.Dispatch, s.serverImp, &reqPackage, &rspPackage, s.withContext)
					if err != nil {
						TLOG.Errorf("Post filter error, No.%v, err: %v", i, err)
					}
				}
			}
			if err != nil {
				TLOG.Errorf("RequestID:%d, Found err: %v", reqPackage.IRequestId, err)
				rspPackage.IRet = 1
				rspPackage.SResultDesc = err.Error()
				if tarsErr, ok := err.(*Error); ok {
					rspPackage.IRet = tarsErr.Code
				}
			}
		}
	}

	// return packet type
	rspPackage.CPacketType = reqPackage.CPacketType
	if ok = current.SetPacketTypeFromContext(ctx, rspPackage.CPacketType); !ok {
		TLOG.Error("SetPacketType in context fail!")
	}

	return s.rsp2Byte(&rspPackage)
}

func (s *Protocol) req2Byte(rsp *requestf.ResponsePacket) []byte {
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
	length := sbuf.Len()
	binary.BigEndian.PutUint32(sbuf.Bytes(), uint32(length))
	return sbuf.Bytes()
}

func (s *Protocol) rsp2Byte(rsp *requestf.ResponsePacket) []byte {
	if rsp.IVersion == basef.TUPVERSION {
		return s.req2Byte(rsp)
	}
	os := codec.NewBuffer()
	rsp.WriteTo(os)
	bs := os.ToBytes()
	sbuf := bytes.NewBuffer(nil)
	sbuf.Write(make([]byte, 4))
	sbuf.Write(bs)
	length := sbuf.Len()
	binary.BigEndian.PutUint32(sbuf.Bytes(), uint32(length))
	return sbuf.Bytes()
}

// ParsePackage parse the []byte according to the tars protocol.
// returns header length and package integrity condition (PackageLess | PackageFull | PackageError)
func (s *Protocol) ParsePackage(buff []byte) (int, int) {
	return protocol.TarsRequest(buff)
}

// InvokeTimeout indicates how to deal with timeout.
func (s *Protocol) InvokeTimeout(pkg []byte) []byte {
	rspPackage := requestf.ResponsePacket{}
	//  invoke timeout need to return IRequestId
	reqPackage := requestf.RequestPacket{}
	is := codec.NewReader(pkg[4:])
	reqPackage.ReadFrom(is)
	rspPackage.IRequestId = reqPackage.IRequestId
	rspPackage.IRet = 1
	rspPackage.SResultDesc = "server invoke timeout"
	return s.rsp2Byte(&rspPackage)
}

// GetCloseMsg return a package to close connection
func (s *Protocol) GetCloseMsg() []byte {
	rspPackage := requestf.ResponsePacket{}
	rspPackage.IVersion = basef.TARSVERSION
	rspPackage.IRequestId = 0
	rspPackage.SResultDesc = reconnectMsg
	return s.rsp2Byte(&rspPackage)
}

// DoClose be called when close connection
func (s *Protocol) DoClose(ctx context.Context) {
	TLOG.Debug("DoClose!")
}
