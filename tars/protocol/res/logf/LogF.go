// Code generated by tars2go 1.2.3, DO NOT EDIT.
// This file was generated from LogF.tars
// Package logf comment
package logf

import (
	"fmt"

	"github.com/TarsCloud/TarsGo/tars/protocol/codec"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = fmt.Errorf
var _ = codec.FromInt8

// LogInfo struct implement
type LogInfo struct {
	Appname           string `json:"appname" tars:"appname,tag:0,require:true"`
	Servername        string `json:"servername" tars:"servername,tag:1,require:true"`
	SFilename         string `json:"sFilename" tars:"sFilename,tag:2,require:true"`
	SFormat           string `json:"sFormat" tars:"sFormat,tag:3,require:true"`
	Setdivision       string `json:"setdivision" tars:"setdivision,tag:4,require:false"`
	BHasSufix         bool   `json:"bHasSufix" tars:"bHasSufix,tag:5,require:false"`
	BHasAppNamePrefix bool   `json:"bHasAppNamePrefix" tars:"bHasAppNamePrefix,tag:6,require:false"`
	BHasSquareBracket bool   `json:"bHasSquareBracket" tars:"bHasSquareBracket,tag:7,require:false"`
	SConcatStr        string `json:"sConcatStr" tars:"sConcatStr,tag:8,require:false"`
	SSepar            string `json:"sSepar" tars:"sSepar,tag:9,require:false"`
	SLogType          string `json:"sLogType" tars:"sLogType,tag:10,require:false"`
}

func (st *LogInfo) ResetDefault() {
	st.BHasSufix = true
	st.BHasAppNamePrefix = true
	st.BHasSquareBracket = false
	st.SConcatStr = "_"
	st.SSepar = "|"
	st.SLogType = ""
}

// ReadFrom reads  from readBuf and put into struct.
func (st *LogInfo) ReadFrom(readBuf *codec.Reader) error {
	var (
		err    error
		length int32
		have   bool
		ty     byte
	)
	st.ResetDefault()

	err = readBuf.ReadString(&st.Appname, 0, true)
	if err != nil {
		return err
	}

	err = readBuf.ReadString(&st.Servername, 1, true)
	if err != nil {
		return err
	}

	err = readBuf.ReadString(&st.SFilename, 2, true)
	if err != nil {
		return err
	}

	err = readBuf.ReadString(&st.SFormat, 3, true)
	if err != nil {
		return err
	}

	err = readBuf.ReadString(&st.Setdivision, 4, false)
	if err != nil {
		return err
	}

	err = readBuf.ReadBool(&st.BHasSufix, 5, false)
	if err != nil {
		return err
	}

	err = readBuf.ReadBool(&st.BHasAppNamePrefix, 6, false)
	if err != nil {
		return err
	}

	err = readBuf.ReadBool(&st.BHasSquareBracket, 7, false)
	if err != nil {
		return err
	}

	err = readBuf.ReadString(&st.SConcatStr, 8, false)
	if err != nil {
		return err
	}

	err = readBuf.ReadString(&st.SSepar, 9, false)
	if err != nil {
		return err
	}

	err = readBuf.ReadString(&st.SLogType, 10, false)
	if err != nil {
		return err
	}

	_ = err
	_ = length
	_ = have
	_ = ty
	return nil
}

// ReadBlock reads struct from the given tag , require or optional.
func (st *LogInfo) ReadBlock(readBuf *codec.Reader, tag byte, require bool) error {
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
			return fmt.Errorf("require LogInfo, but not exist. tag %d", tag)
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
func (st *LogInfo) WriteTo(buf *codec.Buffer) (err error) {
	err = buf.WriteString(st.Appname, 0)
	if err != nil {
		return err
	}

	err = buf.WriteString(st.Servername, 1)
	if err != nil {
		return err
	}

	err = buf.WriteString(st.SFilename, 2)
	if err != nil {
		return err
	}

	err = buf.WriteString(st.SFormat, 3)
	if err != nil {
		return err
	}

	if st.Setdivision != "" {
		err = buf.WriteString(st.Setdivision, 4)
		if err != nil {
			return err
		}
	}

	if st.BHasSufix != true {
		err = buf.WriteBool(st.BHasSufix, 5)
		if err != nil {
			return err
		}
	}

	if st.BHasAppNamePrefix != true {
		err = buf.WriteBool(st.BHasAppNamePrefix, 6)
		if err != nil {
			return err
		}
	}

	if st.BHasSquareBracket != false {
		err = buf.WriteBool(st.BHasSquareBracket, 7)
		if err != nil {
			return err
		}
	}

	if st.SConcatStr != "_" {
		err = buf.WriteString(st.SConcatStr, 8)
		if err != nil {
			return err
		}
	}

	if st.SSepar != "|" {
		err = buf.WriteString(st.SSepar, 9)
		if err != nil {
			return err
		}
	}

	if st.SLogType != "" {
		err = buf.WriteString(st.SLogType, 10)
		if err != nil {
			return err
		}
	}

	return err
}

// WriteBlock encode struct
func (st *LogInfo) WriteBlock(buf *codec.Buffer, tag byte) error {
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
