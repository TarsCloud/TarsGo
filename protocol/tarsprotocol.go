//
// transprotocol.go
// Copyright (C) 2018 toraxie <toraxie@tencent.com>
//
// Distributed under terms of the TENCENT license.
//

package protocol

import (
	"bytes"
	"encoding/binary"
	"tars/protocol/codec"
	"tars/protocol/res/requestf"
)

type TarsProtocol struct {
}

const (
	iMaxLength = 10485760
)

func TafRequest(rev []byte) (int, int) {
	if len(rev) < 4 {
		return 0, PACKAGE_LESS
	}
	iHeaderLen := int(binary.BigEndian.Uint32(rev[0:4]))
	if iHeaderLen < 4 || iHeaderLen > iMaxLength {
		return 0, PACKAGE_ERROR
	}
	if len(rev) < iHeaderLen {
		return 0, PACKAGE_LESS
	}
	return iHeaderLen, PACKAGE_FULL
}

func (self *TarsProtocol) RequestPack(req *requestf.RequestPacket) ([]byte, error) {
	sbuf := bytes.NewBuffer(nil)
	sbuf.Write(make([]byte, 4))
	os := codec.NewBuffer()
	req.WriteTo(os)
	bs := os.ToBytes()
	sbuf.Write(bs)
	len := sbuf.Len()
	binary.BigEndian.PutUint32(sbuf.Bytes(), uint32(len))
	return sbuf.Bytes(), nil

}
func (self *TarsProtocol) ResponseUnpack(pkg []byte) (*requestf.ResponsePacket, error) {
	packet := &requestf.ResponsePacket{}
	err := packet.ReadFrom(codec.NewReader(pkg[4:]))
	return packet, err
}
func (self *TarsProtocol) ParsePackage(rev []byte) (int, int) {
	return TafRequest(rev)
}
