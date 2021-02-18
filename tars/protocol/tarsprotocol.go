package protocol

import (
	"encoding/binary"
	"github.com/TarsCloud/TarsGo/tars/protocol/codec"
	"github.com/TarsCloud/TarsGo/tars/protocol/res/requestf"
)


var maxPackageLength int = 10485760

// SetMaxPackageLength sets the max length of tars packet 
func SetMaxPackageLength(len int) {
	maxPackageLength = len
}

func TarsRequest(rev []byte) (int, int) {
	if len(rev) < 4 {
		return 0, PACKAGE_LESS
	}
	iHeaderLen := int(binary.BigEndian.Uint32(rev[0:4]))
	if iHeaderLen < 4 || iHeaderLen > maxPackageLength {
		return 0, PACKAGE_ERROR
	}
	if len(rev) < iHeaderLen {
		return 0, PACKAGE_LESS
	}
	return iHeaderLen, PACKAGE_FULL
}

type TarsProtocol struct {}

func (p *TarsProtocol) RequestPack(req *requestf.RequestPacket) ([]byte, error) {
	os := codec.NewBuffer()
	err := os.Write_slice_int8(make([]int8, 4))
	if err != nil {
		return nil, err
	}
	err = req.WriteTo(os)
	if err != nil {
		return nil, err
	}
	bs := os.ToBytes()
	l := len(bs)
	binary.BigEndian.PutUint32(bs, uint32(l))
	return bs, nil

}
func (p *TarsProtocol) ResponseUnpack(pkg []byte) (*requestf.ResponsePacket, error) {
	packet := &requestf.ResponsePacket{}
	err := packet.ReadFrom(codec.NewReader(pkg[4:]))
	return packet, err
}
func (p *TarsProtocol) ParsePackage(rev []byte) (int, int) {
	return TarsRequest(rev)
}
