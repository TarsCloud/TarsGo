// Package statf comment
// This file was generated by tars2go 1.1.4
// Generated from StatF.tars
package statf

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	m "github.com/TarsCloud/TarsGo/tars/model"
	"github.com/TarsCloud/TarsGo/tars/protocol/codec"
	"github.com/TarsCloud/TarsGo/tars/protocol/res/basef"
	"github.com/TarsCloud/TarsGo/tars/protocol/res/requestf"
	"github.com/TarsCloud/TarsGo/tars/protocol/tup"
	"github.com/TarsCloud/TarsGo/tars/util/current"
	"github.com/TarsCloud/TarsGo/tars/util/tools"
	"unsafe"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = fmt.Errorf
var _ = codec.FromInt8
var _ = unsafe.Pointer(nil)
var _ = bytes.ErrTooLarge

//StatF struct
type StatF struct {
	s m.Servant
}

//ReportMicMsg is the proxy function for the method defined in the tars file, with the context
func (_obj *StatF) ReportMicMsg(msg map[StatMicMsgHead]StatMicMsgBody, bFromClient bool, _opt ...map[string]string) (ret int32, err error) {

	var length int32
	var have bool
	var ty byte
	_os := codec.NewBuffer()
	err = _os.WriteHead(codec.MAP, 1)
	if err != nil {
		return ret, err
	}

	err = _os.Write_int32(int32(len(msg)), 0)
	if err != nil {
		return ret, err
	}

	for k0, v0 := range msg {

		err = k0.WriteBlock(_os, 0)
		if err != nil {
			return ret, err
		}

		err = v0.WriteBlock(_os, 1)
		if err != nil {
			return ret, err
		}

	}

	err = _os.Write_bool(bFromClient, 2)
	if err != nil {
		return ret, err
	}

	var _status map[string]string
	var _context map[string]string
	if len(_opt) == 1 {
		_context = _opt[0]
	} else if len(_opt) == 2 {
		_context = _opt[0]
		_status = _opt[1]
	}
	_resp := new(requestf.ResponsePacket)
	tarsCtx := context.Background()

	err = _obj.s.Tars_invoke(tarsCtx, 0, "reportMicMsg", _os.ToBytes(), _status, _context, _resp)
	if err != nil {
		return ret, err
	}

	_is := codec.NewReader(tools.Int8ToByte(_resp.SBuffer))
	err = _is.Read_int32(&ret, 0, true)
	if err != nil {
		return ret, err
	}

	if len(_opt) == 1 {
		for k := range _context {
			delete(_context, k)
		}
		for k, v := range _resp.Context {
			_context[k] = v
		}
	} else if len(_opt) == 2 {
		for k := range _context {
			delete(_context, k)
		}
		for k, v := range _resp.Context {
			_context[k] = v
		}
		for k := range _status {
			delete(_status, k)
		}
		for k, v := range _resp.Status {
			_status[k] = v
		}

	}
	_ = length
	_ = have
	_ = ty
	return ret, nil
}

//ReportMicMsgWithContext is the proxy function for the method defined in the tars file, with the context
func (_obj *StatF) ReportMicMsgWithContext(tarsCtx context.Context, msg map[StatMicMsgHead]StatMicMsgBody, bFromClient bool, _opt ...map[string]string) (ret int32, err error) {

	var length int32
	var have bool
	var ty byte
	_os := codec.NewBuffer()
	err = _os.WriteHead(codec.MAP, 1)
	if err != nil {
		return ret, err
	}

	err = _os.Write_int32(int32(len(msg)), 0)
	if err != nil {
		return ret, err
	}

	for k1, v1 := range msg {

		err = k1.WriteBlock(_os, 0)
		if err != nil {
			return ret, err
		}

		err = v1.WriteBlock(_os, 1)
		if err != nil {
			return ret, err
		}

	}

	err = _os.Write_bool(bFromClient, 2)
	if err != nil {
		return ret, err
	}

	var _status map[string]string
	var _context map[string]string
	if len(_opt) == 1 {
		_context = _opt[0]
	} else if len(_opt) == 2 {
		_context = _opt[0]
		_status = _opt[1]
	}
	_resp := new(requestf.ResponsePacket)

	err = _obj.s.Tars_invoke(tarsCtx, 0, "reportMicMsg", _os.ToBytes(), _status, _context, _resp)
	if err != nil {
		return ret, err
	}

	_is := codec.NewReader(tools.Int8ToByte(_resp.SBuffer))
	err = _is.Read_int32(&ret, 0, true)
	if err != nil {
		return ret, err
	}

	if len(_opt) == 1 {
		for k := range _context {
			delete(_context, k)
		}
		for k, v := range _resp.Context {
			_context[k] = v
		}
	} else if len(_opt) == 2 {
		for k := range _context {
			delete(_context, k)
		}
		for k, v := range _resp.Context {
			_context[k] = v
		}
		for k := range _status {
			delete(_status, k)
		}
		for k, v := range _resp.Status {
			_status[k] = v
		}

	}
	_ = length
	_ = have
	_ = ty
	return ret, nil
}

//ReportMicMsgOneWayWithContext is the proxy function for the method defined in the tars file, with the context
func (_obj *StatF) ReportMicMsgOneWayWithContext(tarsCtx context.Context, msg map[StatMicMsgHead]StatMicMsgBody, bFromClient bool, _opt ...map[string]string) (ret int32, err error) {

	var length int32
	var have bool
	var ty byte
	_os := codec.NewBuffer()
	err = _os.WriteHead(codec.MAP, 1)
	if err != nil {
		return ret, err
	}

	err = _os.Write_int32(int32(len(msg)), 0)
	if err != nil {
		return ret, err
	}

	for k2, v2 := range msg {

		err = k2.WriteBlock(_os, 0)
		if err != nil {
			return ret, err
		}

		err = v2.WriteBlock(_os, 1)
		if err != nil {
			return ret, err
		}

	}

	err = _os.Write_bool(bFromClient, 2)
	if err != nil {
		return ret, err
	}

	var _status map[string]string
	var _context map[string]string
	if len(_opt) == 1 {
		_context = _opt[0]
	} else if len(_opt) == 2 {
		_context = _opt[0]
		_status = _opt[1]
	}
	_resp := new(requestf.ResponsePacket)

	err = _obj.s.Tars_invoke(tarsCtx, 1, "reportMicMsg", _os.ToBytes(), _status, _context, _resp)
	if err != nil {
		return ret, err
	}

	if len(_opt) == 1 {
		for k := range _context {
			delete(_context, k)
		}
		for k, v := range _resp.Context {
			_context[k] = v
		}
	} else if len(_opt) == 2 {
		for k := range _context {
			delete(_context, k)
		}
		for k, v := range _resp.Context {
			_context[k] = v
		}
		for k := range _status {
			delete(_status, k)
		}
		for k, v := range _resp.Status {
			_status[k] = v
		}

	}
	_ = length
	_ = have
	_ = ty
	return ret, nil
}

//ReportSampleMsg is the proxy function for the method defined in the tars file, with the context
func (_obj *StatF) ReportSampleMsg(msg []StatSampleMsg, _opt ...map[string]string) (ret int32, err error) {

	var length int32
	var have bool
	var ty byte
	_os := codec.NewBuffer()
	err = _os.WriteHead(codec.LIST, 1)
	if err != nil {
		return ret, err
	}

	err = _os.Write_int32(int32(len(msg)), 0)
	if err != nil {
		return ret, err
	}

	for _, v := range msg {

		err = v.WriteBlock(_os, 0)
		if err != nil {
			return ret, err
		}

	}

	var _status map[string]string
	var _context map[string]string
	if len(_opt) == 1 {
		_context = _opt[0]
	} else if len(_opt) == 2 {
		_context = _opt[0]
		_status = _opt[1]
	}
	_resp := new(requestf.ResponsePacket)
	tarsCtx := context.Background()

	err = _obj.s.Tars_invoke(tarsCtx, 0, "reportSampleMsg", _os.ToBytes(), _status, _context, _resp)
	if err != nil {
		return ret, err
	}

	_is := codec.NewReader(tools.Int8ToByte(_resp.SBuffer))
	err = _is.Read_int32(&ret, 0, true)
	if err != nil {
		return ret, err
	}

	if len(_opt) == 1 {
		for k := range _context {
			delete(_context, k)
		}
		for k, v := range _resp.Context {
			_context[k] = v
		}
	} else if len(_opt) == 2 {
		for k := range _context {
			delete(_context, k)
		}
		for k, v := range _resp.Context {
			_context[k] = v
		}
		for k := range _status {
			delete(_status, k)
		}
		for k, v := range _resp.Status {
			_status[k] = v
		}

	}
	_ = length
	_ = have
	_ = ty
	return ret, nil
}

//ReportSampleMsgWithContext is the proxy function for the method defined in the tars file, with the context
func (_obj *StatF) ReportSampleMsgWithContext(tarsCtx context.Context, msg []StatSampleMsg, _opt ...map[string]string) (ret int32, err error) {

	var length int32
	var have bool
	var ty byte
	_os := codec.NewBuffer()
	err = _os.WriteHead(codec.LIST, 1)
	if err != nil {
		return ret, err
	}

	err = _os.Write_int32(int32(len(msg)), 0)
	if err != nil {
		return ret, err
	}

	for _, v := range msg {

		err = v.WriteBlock(_os, 0)
		if err != nil {
			return ret, err
		}

	}

	var _status map[string]string
	var _context map[string]string
	if len(_opt) == 1 {
		_context = _opt[0]
	} else if len(_opt) == 2 {
		_context = _opt[0]
		_status = _opt[1]
	}
	_resp := new(requestf.ResponsePacket)

	err = _obj.s.Tars_invoke(tarsCtx, 0, "reportSampleMsg", _os.ToBytes(), _status, _context, _resp)
	if err != nil {
		return ret, err
	}

	_is := codec.NewReader(tools.Int8ToByte(_resp.SBuffer))
	err = _is.Read_int32(&ret, 0, true)
	if err != nil {
		return ret, err
	}

	if len(_opt) == 1 {
		for k := range _context {
			delete(_context, k)
		}
		for k, v := range _resp.Context {
			_context[k] = v
		}
	} else if len(_opt) == 2 {
		for k := range _context {
			delete(_context, k)
		}
		for k, v := range _resp.Context {
			_context[k] = v
		}
		for k := range _status {
			delete(_status, k)
		}
		for k, v := range _resp.Status {
			_status[k] = v
		}

	}
	_ = length
	_ = have
	_ = ty
	return ret, nil
}

//ReportSampleMsgOneWayWithContext is the proxy function for the method defined in the tars file, with the context
func (_obj *StatF) ReportSampleMsgOneWayWithContext(tarsCtx context.Context, msg []StatSampleMsg, _opt ...map[string]string) (ret int32, err error) {

	var length int32
	var have bool
	var ty byte
	_os := codec.NewBuffer()
	err = _os.WriteHead(codec.LIST, 1)
	if err != nil {
		return ret, err
	}

	err = _os.Write_int32(int32(len(msg)), 0)
	if err != nil {
		return ret, err
	}

	for _, v := range msg {

		err = v.WriteBlock(_os, 0)
		if err != nil {
			return ret, err
		}

	}

	var _status map[string]string
	var _context map[string]string
	if len(_opt) == 1 {
		_context = _opt[0]
	} else if len(_opt) == 2 {
		_context = _opt[0]
		_status = _opt[1]
	}
	_resp := new(requestf.ResponsePacket)

	err = _obj.s.Tars_invoke(tarsCtx, 1, "reportSampleMsg", _os.ToBytes(), _status, _context, _resp)
	if err != nil {
		return ret, err
	}

	if len(_opt) == 1 {
		for k := range _context {
			delete(_context, k)
		}
		for k, v := range _resp.Context {
			_context[k] = v
		}
	} else if len(_opt) == 2 {
		for k := range _context {
			delete(_context, k)
		}
		for k, v := range _resp.Context {
			_context[k] = v
		}
		for k := range _status {
			delete(_status, k)
		}
		for k, v := range _resp.Status {
			_status[k] = v
		}

	}
	_ = length
	_ = have
	_ = ty
	return ret, nil
}

//SetServant sets servant for the service.
func (_obj *StatF) SetServant(s m.Servant) {
	_obj.s = s
}

//TarsSetTimeout sets the timeout for the servant which is in ms.
func (_obj *StatF) TarsSetTimeout(t int) {
	_obj.s.TarsSetTimeout(t)
}

//TarsSetProtocol sets the protocol for the servant.
func (_obj *StatF) TarsSetProtocol(p m.Protocol) {
	_obj.s.TarsSetProtocol(p)
}

type _impStatF interface {
	ReportMicMsg(msg map[StatMicMsgHead]StatMicMsgBody, bFromClient bool) (ret int32, err error)
	ReportSampleMsg(msg []StatSampleMsg) (ret int32, err error)
}
type _impStatFWithContext interface {
	ReportMicMsg(tarsCtx context.Context, msg map[StatMicMsgHead]StatMicMsgBody, bFromClient bool) (ret int32, err error)
	ReportSampleMsg(tarsCtx context.Context, msg []StatSampleMsg) (ret int32, err error)
}

// Dispatch is used to call the server side implemnet for the method defined in the tars file. _withContext shows using context or not.
func (_obj *StatF) Dispatch(tarsCtx context.Context, _val interface{}, tarsReq *requestf.RequestPacket, tarsResp *requestf.ResponsePacket, _withContext bool) (err error) {
	var length int32
	var have bool
	var ty byte
	_is := codec.NewReader(tools.Int8ToByte(tarsReq.SBuffer))
	_os := codec.NewBuffer()
	switch tarsReq.SFuncName {
	case "reportMicMsg":
		var msg map[StatMicMsgHead]StatMicMsgBody
		msg = make(map[StatMicMsgHead]StatMicMsgBody)
		var bFromClient bool

		if tarsReq.IVersion == basef.TARSVERSION {

			_, err = _is.SkipTo(codec.MAP, 1, true)
			if err != nil {
				return err
			}

			err = _is.Read_int32(&length, 0, true)
			if err != nil {
				return err
			}

			msg = make(map[StatMicMsgHead]StatMicMsgBody)
			for i3, e3 := int32(0), length; i3 < e3; i3++ {
				var k3 StatMicMsgHead
				var v3 StatMicMsgBody

				err = k3.ReadBlock(_is, 0, false)
				if err != nil {
					return err
				}

				err = v3.ReadBlock(_is, 1, false)
				if err != nil {
					return err
				}

				msg[k3] = v3
			}

			err = _is.Read_bool(&bFromClient, 2, true)
			if err != nil {
				return err
			}

		} else if tarsReq.IVersion == basef.TUPVERSION {
			_reqTup_ := tup.NewUniAttribute()
			_reqTup_.Decode(_is)

			var _tupBuffer_ []byte

			_reqTup_.GetBuffer("msg", &_tupBuffer_)
			_is.Reset(_tupBuffer_)
			_, err = _is.SkipTo(codec.MAP, 0, true)
			if err != nil {
				return err
			}

			err = _is.Read_int32(&length, 0, true)
			if err != nil {
				return err
			}

			msg = make(map[StatMicMsgHead]StatMicMsgBody)
			for i4, e4 := int32(0), length; i4 < e4; i4++ {
				var k4 StatMicMsgHead
				var v4 StatMicMsgBody

				err = k4.ReadBlock(_is, 0, false)
				if err != nil {
					return err
				}

				err = v4.ReadBlock(_is, 1, false)
				if err != nil {
					return err
				}

				msg[k4] = v4
			}

			_reqTup_.GetBuffer("bFromClient", &_tupBuffer_)
			_is.Reset(_tupBuffer_)
			err = _is.Read_bool(&bFromClient, 0, true)
			if err != nil {
				return err
			}

		} else if tarsReq.IVersion == basef.JSONVERSION {
			var _jsonDat_ map[string]interface{}
			_decoder_ := json.NewDecoder(bytes.NewReader(_is.ToBytes()))
			_decoder_.UseNumber()
			err = _decoder_.Decode(&_jsonDat_)
			if err != nil {
				return fmt.Errorf("decode reqpacket failed, error: %+v", err)
			}
			{
				_jsonStr_, _ := json.Marshal(_jsonDat_["msg"])
				if err = json.Unmarshal([]byte(_jsonStr_), &msg); err != nil {
					return err
				}
			}
			{
				_jsonStr_, _ := json.Marshal(_jsonDat_["bFromClient"])
				if err = json.Unmarshal([]byte(_jsonStr_), &bFromClient); err != nil {
					return err
				}
			}

		} else {
			err = fmt.Errorf("decode reqpacket fail, error version: %d", tarsReq.IVersion)
			return err
		}

		var _funRet_ int32
		if !_withContext {
			_imp := _val.(_impStatF)
			_funRet_, err = _imp.ReportMicMsg(msg, bFromClient)
		} else {
			_imp := _val.(_impStatFWithContext)
			_funRet_, err = _imp.ReportMicMsg(tarsCtx, msg, bFromClient)
		}

		if err != nil {
			return err
		}

		if tarsReq.IVersion == basef.TARSVERSION {
			_os.Reset()

			err = _os.Write_int32(_funRet_, 0)
			if err != nil {
				return err
			}

		} else if tarsReq.IVersion == basef.TUPVERSION {
			_tupRsp_ := tup.NewUniAttribute()

			err = _os.Write_int32(_funRet_, 0)
			if err != nil {
				return err
			}

			_tupRsp_.PutBuffer("", _os.ToBytes())
			_tupRsp_.PutBuffer("tars_ret", _os.ToBytes())

			_os.Reset()
			err = _tupRsp_.Encode(_os)
			if err != nil {
				return err
			}
		} else if tarsReq.IVersion == basef.JSONVERSION {
			_rspJson_ := map[string]interface{}{}
			_rspJson_["tars_ret"] = _funRet_

			var _rspByte_ []byte
			if _rspByte_, err = json.Marshal(_rspJson_); err != nil {
				return err
			}

			_os.Reset()
			err = _os.Write_slice_uint8(_rspByte_)
			if err != nil {
				return err
			}
		}
	case "reportSampleMsg":
		var msg []StatSampleMsg

		if tarsReq.IVersion == basef.TARSVERSION {

			_, ty, err = _is.SkipToNoCheck(1, true)
			if err != nil {
				return err
			}

			if ty == codec.LIST {
				err = _is.Read_int32(&length, 0, true)
				if err != nil {
					return err
				}

				msg = make([]StatSampleMsg, length)
				for i5, e5 := int32(0), length; i5 < e5; i5++ {

					err = msg[i5].ReadBlock(_is, 0, false)
					if err != nil {
						return err
					}

				}
			} else if ty == codec.SIMPLE_LIST {
				err = fmt.Errorf("not support simple_list type")
				if err != nil {
					return err
				}

			} else {
				err = fmt.Errorf("require vector, but not")
				if err != nil {
					return err
				}

			}
		} else if tarsReq.IVersion == basef.TUPVERSION {
			_reqTup_ := tup.NewUniAttribute()
			_reqTup_.Decode(_is)

			var _tupBuffer_ []byte

			_reqTup_.GetBuffer("msg", &_tupBuffer_)
			_is.Reset(_tupBuffer_)
			_, ty, err = _is.SkipToNoCheck(0, true)
			if err != nil {
				return err
			}

			if ty == codec.LIST {
				err = _is.Read_int32(&length, 0, true)
				if err != nil {
					return err
				}

				msg = make([]StatSampleMsg, length)
				for i6, e6 := int32(0), length; i6 < e6; i6++ {

					err = msg[i6].ReadBlock(_is, 0, false)
					if err != nil {
						return err
					}

				}
			} else if ty == codec.SIMPLE_LIST {
				err = fmt.Errorf("not support simple_list type")
				if err != nil {
					return err
				}

			} else {
				err = fmt.Errorf("require vector, but not")
				if err != nil {
					return err
				}

			}
		} else if tarsReq.IVersion == basef.JSONVERSION {
			var _jsonDat_ map[string]interface{}
			_decoder_ := json.NewDecoder(bytes.NewReader(_is.ToBytes()))
			_decoder_.UseNumber()
			err = _decoder_.Decode(&_jsonDat_)
			if err != nil {
				return fmt.Errorf("decode reqpacket failed, error: %+v", err)
			}
			{
				_jsonStr_, _ := json.Marshal(_jsonDat_["msg"])
				if err = json.Unmarshal([]byte(_jsonStr_), &msg); err != nil {
					return err
				}
			}

		} else {
			err = fmt.Errorf("decode reqpacket fail, error version: %d", tarsReq.IVersion)
			return err
		}

		var _funRet_ int32
		if !_withContext {
			_imp := _val.(_impStatF)
			_funRet_, err = _imp.ReportSampleMsg(msg)
		} else {
			_imp := _val.(_impStatFWithContext)
			_funRet_, err = _imp.ReportSampleMsg(tarsCtx, msg)
		}

		if err != nil {
			return err
		}

		if tarsReq.IVersion == basef.TARSVERSION {
			_os.Reset()

			err = _os.Write_int32(_funRet_, 0)
			if err != nil {
				return err
			}

		} else if tarsReq.IVersion == basef.TUPVERSION {
			_tupRsp_ := tup.NewUniAttribute()

			err = _os.Write_int32(_funRet_, 0)
			if err != nil {
				return err
			}

			_tupRsp_.PutBuffer("", _os.ToBytes())
			_tupRsp_.PutBuffer("tars_ret", _os.ToBytes())

			_os.Reset()
			err = _tupRsp_.Encode(_os)
			if err != nil {
				return err
			}
		} else if tarsReq.IVersion == basef.JSONVERSION {
			_rspJson_ := map[string]interface{}{}
			_rspJson_["tars_ret"] = _funRet_

			var _rspByte_ []byte
			if _rspByte_, err = json.Marshal(_rspJson_); err != nil {
				return err
			}

			_os.Reset()
			err = _os.Write_slice_uint8(_rspByte_)
			if err != nil {
				return err
			}
		}

	default:
		return fmt.Errorf("func mismatch")
	}
	var _status map[string]string
	s, ok := current.GetResponseStatus(tarsCtx)
	if ok && s != nil {
		_status = s
	}
	var _context map[string]string
	c, ok := current.GetResponseContext(tarsCtx)
	if ok && c != nil {
		_context = c
	}
	*tarsResp = requestf.ResponsePacket{
		IVersion:     tarsReq.IVersion,
		CPacketType:  0,
		IRequestId:   tarsReq.IRequestId,
		IMessageType: 0,
		IRet:         0,
		SBuffer:      tools.ByteToInt8(_os.ToBytes()),
		Status:       _status,
		SResultDesc:  "",
		Context:      _context,
	}

	_ = _is
	_ = _os
	_ = length
	_ = have
	_ = ty
	return nil
}
