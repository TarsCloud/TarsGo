// Package requestf comment
// This file was generated by tars2go 2.0.0
// Generated from RequestF.tars
package requestf

import (
	"fmt"

	"github.com/TarsCloud/TarsGo/tars/protocol/codec"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = fmt.Errorf
var _ = codec.FromInt8

// RequestPacket struct implement
type RequestPacket struct {
	IVersion     int16             `json:"iVersion"`
	CPacketType  int8              `json:"cPacketType"`
	IMessageType int32             `json:"iMessageType"`
	IRequestId   int32             `json:"iRequestId"`
	SServantName string            `json:"sServantName"`
	SFuncName    string            `json:"sFuncName"`
	SBuffer      []int8            `json:"sBuffer"`
	ITimeout     int32             `json:"iTimeout"`
	Context      map[string]string `json:"context"`
	Status       map[string]string `json:"status"`
}

func (st *RequestPacket) ResetDefault() {
	st.CPacketType = 0
	st.IMessageType = 0
	st.SServantName = ""
	st.SFuncName = ""
	st.ITimeout = 0
}

// ReadFrom reads  from readBuf and put into struct.
func (st *RequestPacket) ReadFrom(readBuf *codec.Reader) error {
	var (
		err    error
		length int32
		have   bool
		ty     byte
	)
	st.ResetDefault()

	err = readBuf.ReadInt16(&st.IVersion, 1, true)
	if err != nil {
		return err
	}

	err = readBuf.ReadInt8(&st.CPacketType, 2, true)
	if err != nil {
		return err
	}

	err = readBuf.ReadInt32(&st.IMessageType, 3, true)
	if err != nil {
		return err
	}

	err = readBuf.ReadInt32(&st.IRequestId, 4, true)
	if err != nil {
		return err
	}

	err = readBuf.ReadString(&st.SServantName, 5, true)
	if err != nil {
		return err
	}

	err = readBuf.ReadString(&st.SFuncName, 6, true)
	if err != nil {
		return err
	}

	_, ty, err = readBuf.SkipToNoCheck(7, true)
	if err != nil {
		return err
	}

	if ty == codec.LIST {
		err = readBuf.ReadInt32(&length, 0, true)
		if err != nil {
			return err
		}

		st.SBuffer = make([]int8, length)
		for i0, e0 := int32(0), length; i0 < e0; i0++ {

			err = readBuf.ReadInt8(&st.SBuffer[i0], 0, false)
			if err != nil {
				return err
			}

		}
	} else if ty == codec.SimpleList {

		_, err = readBuf.SkipTo(codec.BYTE, 0, true)
		if err != nil {
			return err
		}

		err = readBuf.ReadInt32(&length, 0, true)
		if err != nil {
			return err
		}

		err = readBuf.ReadSliceInt8(&st.SBuffer, length, true)
		if err != nil {
			return err
		}

	} else {
		err = fmt.Errorf("require vector, but not")
		if err != nil {
			return err
		}

	}

	err = readBuf.ReadInt32(&st.ITimeout, 8, true)
	if err != nil {
		return err
	}

	_, err = readBuf.SkipTo(codec.MAP, 9, true)
	if err != nil {
		return err
	}

	err = readBuf.ReadInt32(&length, 0, true)
	if err != nil {
		return err
	}

	st.Context = make(map[string]string)
	for i1, e1 := int32(0), length; i1 < e1; i1++ {
		var k1 string
		var v1 string

		err = readBuf.ReadString(&k1, 0, false)
		if err != nil {
			return err
		}

		err = readBuf.ReadString(&v1, 1, false)
		if err != nil {
			return err
		}

		st.Context[k1] = v1
	}

	_, err = readBuf.SkipTo(codec.MAP, 10, true)
	if err != nil {
		return err
	}

	err = readBuf.ReadInt32(&length, 0, true)
	if err != nil {
		return err
	}

	st.Status = make(map[string]string)
	for i2, e2 := int32(0), length; i2 < e2; i2++ {
		var k2 string
		var v2 string

		err = readBuf.ReadString(&k2, 0, false)
		if err != nil {
			return err
		}

		err = readBuf.ReadString(&v2, 1, false)
		if err != nil {
			return err
		}

		st.Status[k2] = v2
	}

	_ = err
	_ = length
	_ = have
	_ = ty
	return nil
}

// ReadBlock reads struct from the given tag , require or optional.
func (st *RequestPacket) ReadBlock(readBuf *codec.Reader, tag byte, require bool) error {
	var (
		err  error
		have bool
	)
	st.ResetDefault()

	have, err = readBuf.SkipTo(codec.StructBegin, tag, require)
	if err != nil {
		return err
	}
	if !have {
		if require {
			return fmt.Errorf("require RequestPacket, but not exist. tag %d", tag)
		}
		return nil
	}

	err = st.ReadFrom(readBuf)
	if err != nil {
		return err
	}

	err = readBuf.SkipToStructEnd()
	if err != nil {
		return err
	}
	_ = have
	return nil
}

// WriteTo encode struct to buffer
func (st *RequestPacket) WriteTo(buf *codec.Buffer) error {
	var err error

	err = buf.WriteInt16(st.IVersion, 1)
	if err != nil {
		return err
	}

	err = buf.WriteInt8(st.CPacketType, 2)
	if err != nil {
		return err
	}

	err = buf.WriteInt32(st.IMessageType, 3)
	if err != nil {
		return err
	}

	err = buf.WriteInt32(st.IRequestId, 4)
	if err != nil {
		return err
	}

	err = buf.WriteString(st.SServantName, 5)
	if err != nil {
		return err
	}

	err = buf.WriteString(st.SFuncName, 6)
	if err != nil {
		return err
	}

	err = buf.WriteHead(codec.SimpleList, 7)
	if err != nil {
		return err
	}

	err = buf.WriteHead(codec.BYTE, 0)
	if err != nil {
		return err
	}

	err = buf.WriteInt32(int32(len(st.SBuffer)), 0)
	if err != nil {
		return err
	}

	err = buf.WriteSliceInt8(st.SBuffer)
	if err != nil {
		return err
	}

	err = buf.WriteInt32(st.ITimeout, 8)
	if err != nil {
		return err
	}

	err = buf.WriteHead(codec.MAP, 9)
	if err != nil {
		return err
	}

	err = buf.WriteInt32(int32(len(st.Context)), 0)
	if err != nil {
		return err
	}

	for k3, v3 := range st.Context {

		err = buf.WriteString(k3, 0)
		if err != nil {
			return err
		}

		err = buf.WriteString(v3, 1)
		if err != nil {
			return err
		}

	}

	err = buf.WriteHead(codec.MAP, 10)
	if err != nil {
		return err
	}

	err = buf.WriteInt32(int32(len(st.Status)), 0)
	if err != nil {
		return err
	}

	for k4, v4 := range st.Status {

		err = buf.WriteString(k4, 0)
		if err != nil {
			return err
		}

		err = buf.WriteString(v4, 1)
		if err != nil {
			return err
		}

	}

	_ = err

	return nil
}

// WriteBlock encode struct
func (st *RequestPacket) WriteBlock(buf *codec.Buffer, tag byte) error {
	var err error
	err = buf.WriteHead(codec.StructBegin, tag)
	if err != nil {
		return err
	}

	err = st.WriteTo(buf)
	if err != nil {
		return err
	}

	err = buf.WriteHead(codec.StructEnd, 0)
	if err != nil {
		return err
	}
	return nil
}

// ResponsePacket struct implement
type ResponsePacket struct {
	IVersion     int16             `json:"iVersion"`
	CPacketType  int8              `json:"cPacketType"`
	IRequestId   int32             `json:"iRequestId"`
	IMessageType int32             `json:"iMessageType"`
	IRet         int32             `json:"iRet"`
	SBuffer      []int8            `json:"sBuffer"`
	Status       map[string]string `json:"status"`
	SResultDesc  string            `json:"sResultDesc"`
	Context      map[string]string `json:"context"`
}

func (st *ResponsePacket) ResetDefault() {
	st.CPacketType = 0
	st.IMessageType = 0
	st.IRet = 0
}

// ReadFrom reads  from readBuf and put into struct.
func (st *ResponsePacket) ReadFrom(readBuf *codec.Reader) error {
	var (
		err    error
		length int32
		have   bool
		ty     byte
	)
	st.ResetDefault()

	err = readBuf.ReadInt16(&st.IVersion, 1, true)
	if err != nil {
		return err
	}

	err = readBuf.ReadInt8(&st.CPacketType, 2, true)
	if err != nil {
		return err
	}

	err = readBuf.ReadInt32(&st.IRequestId, 3, true)
	if err != nil {
		return err
	}

	err = readBuf.ReadInt32(&st.IMessageType, 4, true)
	if err != nil {
		return err
	}

	err = readBuf.ReadInt32(&st.IRet, 5, true)
	if err != nil {
		return err
	}

	_, ty, err = readBuf.SkipToNoCheck(6, true)
	if err != nil {
		return err
	}

	if ty == codec.LIST {
		err = readBuf.ReadInt32(&length, 0, true)
		if err != nil {
			return err
		}

		st.SBuffer = make([]int8, length)
		for i0, e0 := int32(0), length; i0 < e0; i0++ {

			err = readBuf.ReadInt8(&st.SBuffer[i0], 0, false)
			if err != nil {
				return err
			}

		}
	} else if ty == codec.SimpleList {

		_, err = readBuf.SkipTo(codec.BYTE, 0, true)
		if err != nil {
			return err
		}

		err = readBuf.ReadInt32(&length, 0, true)
		if err != nil {
			return err
		}

		err = readBuf.ReadSliceInt8(&st.SBuffer, length, true)
		if err != nil {
			return err
		}

	} else {
		err = fmt.Errorf("require vector, but not")
		if err != nil {
			return err
		}

	}

	_, err = readBuf.SkipTo(codec.MAP, 7, true)
	if err != nil {
		return err
	}

	err = readBuf.ReadInt32(&length, 0, true)
	if err != nil {
		return err
	}

	st.Status = make(map[string]string)
	for i1, e1 := int32(0), length; i1 < e1; i1++ {
		var k1 string
		var v1 string

		err = readBuf.ReadString(&k1, 0, false)
		if err != nil {
			return err
		}

		err = readBuf.ReadString(&v1, 1, false)
		if err != nil {
			return err
		}

		st.Status[k1] = v1
	}

	err = readBuf.ReadString(&st.SResultDesc, 8, false)
	if err != nil {
		return err
	}

	have, err = readBuf.SkipTo(codec.MAP, 9, false)
	if err != nil {
		return err
	}

	if have {
		err = readBuf.ReadInt32(&length, 0, true)
		if err != nil {
			return err
		}

		st.Context = make(map[string]string)
		for i2, e2 := int32(0), length; i2 < e2; i2++ {
			var k2 string
			var v2 string

			err = readBuf.ReadString(&k2, 0, false)
			if err != nil {
				return err
			}

			err = readBuf.ReadString(&v2, 1, false)
			if err != nil {
				return err
			}

			st.Context[k2] = v2
		}
	}

	_ = err
	_ = length
	_ = have
	_ = ty
	return nil
}

// ReadBlock reads struct from the given tag , require or optional.
func (st *ResponsePacket) ReadBlock(readBuf *codec.Reader, tag byte, require bool) error {
	var (
		err  error
		have bool
	)
	st.ResetDefault()

	have, err = readBuf.SkipTo(codec.StructBegin, tag, require)
	if err != nil {
		return err
	}
	if !have {
		if require {
			return fmt.Errorf("require ResponsePacket, but not exist. tag %d", tag)
		}
		return nil
	}

	err = st.ReadFrom(readBuf)
	if err != nil {
		return err
	}

	err = readBuf.SkipToStructEnd()
	if err != nil {
		return err
	}
	_ = have
	return nil
}

// WriteTo encode struct to buffer
func (st *ResponsePacket) WriteTo(buf *codec.Buffer) error {
	var err error

	err = buf.WriteInt16(st.IVersion, 1)
	if err != nil {
		return err
	}

	err = buf.WriteInt8(st.CPacketType, 2)
	if err != nil {
		return err
	}

	err = buf.WriteInt32(st.IRequestId, 3)
	if err != nil {
		return err
	}

	err = buf.WriteInt32(st.IMessageType, 4)
	if err != nil {
		return err
	}

	err = buf.WriteInt32(st.IRet, 5)
	if err != nil {
		return err
	}

	err = buf.WriteHead(codec.SimpleList, 6)
	if err != nil {
		return err
	}

	err = buf.WriteHead(codec.BYTE, 0)
	if err != nil {
		return err
	}

	err = buf.WriteInt32(int32(len(st.SBuffer)), 0)
	if err != nil {
		return err
	}

	err = buf.WriteSliceInt8(st.SBuffer)
	if err != nil {
		return err
	}

	err = buf.WriteHead(codec.MAP, 7)
	if err != nil {
		return err
	}

	err = buf.WriteInt32(int32(len(st.Status)), 0)
	if err != nil {
		return err
	}

	for k3, v3 := range st.Status {

		err = buf.WriteString(k3, 0)
		if err != nil {
			return err
		}

		err = buf.WriteString(v3, 1)
		if err != nil {
			return err
		}

	}

	err = buf.WriteString(st.SResultDesc, 8)
	if err != nil {
		return err
	}

	err = buf.WriteHead(codec.MAP, 9)
	if err != nil {
		return err
	}

	err = buf.WriteInt32(int32(len(st.Context)), 0)
	if err != nil {
		return err
	}

	for k4, v4 := range st.Context {

		err = buf.WriteString(k4, 0)
		if err != nil {
			return err
		}

		err = buf.WriteString(v4, 1)
		if err != nil {
			return err
		}

	}

	_ = err

	return nil
}

// WriteBlock encode struct
func (st *ResponsePacket) WriteBlock(buf *codec.Buffer, tag byte) error {
	var err error
	err = buf.WriteHead(codec.StructBegin, tag)
	if err != nil {
		return err
	}

	err = st.WriteTo(buf)
	if err != nil {
		return err
	}

	err = buf.WriteHead(codec.StructEnd, 0)
	if err != nil {
		return err
	}
	return nil
}
