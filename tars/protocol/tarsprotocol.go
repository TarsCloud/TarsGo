package protocol

import (
	"bytes"
	"encoding/binary"

	"github.com/TarsCloud/TarsGo/tars/protocol/codec"
	"github.com/TarsCloud/TarsGo/tars/protocol/res/requestf"
)


func TarsRequest(rev []byte, maxLen int) (int, int) {
	if len(rev) < 4 {
		return 0, PACKAGE_LESS
	}
	iHeaderLen := int(binary.BigEndian.Uint32(rev[0:4]))
	if iHeaderLen < 4 || iHeaderLen > maxLen {
		return 0, PACKAGE_ERROR
	}
	if len(rev) < iHeaderLen {
		return 0, PACKAGE_LESS
	}
	return iHeaderLen, PACKAGE_FULL
}

type TarsProtocol struct {
	MaxPackageLength int
}

func (p *TarsProtocol) RequestPack(req *requestf.RequestPacket) ([]byte, error) {
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
func (p *TarsProtocol) ResponseUnpack(pkg []byte) (*requestf.ResponsePacket, error) {
	packet := &requestf.ResponsePacket{}
	err := packet.ReadFrom(codec.NewReader(pkg[4:]))
	return packet, err
}
func (p *TarsProtocol) ParsePackage(rev []byte) (int, int) {
	return TarsRequest(rev, p.MaxPackageLength)
}
