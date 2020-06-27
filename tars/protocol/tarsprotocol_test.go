package protocol

import (
	"bytes"
	"encoding/binary"
	"github.com/TarsCloud/TarsGo/tars/protocol/codec"
	"github.com/TarsCloud/TarsGo/tars/protocol/res/basef"
	"github.com/TarsCloud/TarsGo/tars/protocol/res/requestf"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSetMaxPackageLength(t *testing.T) {
	type args struct {
		len int
	}
	tests := []int {1,2,3,4,5,10,100,1000,10230,100042, 1000523}
	for _, tt := range tests {
		t.Run("normal", func(t *testing.T) {
			SetMaxPackageLength(tt)
			if maxPackageLength != tt {
				t.Errorf("SetMaxPackageLength failed. want:%v, got:%v\n", tt, maxPackageLength)
			}
		})
	}
}

func TestTarsProtocol_ParsePackage(t *testing.T) {
	type fields struct {
		MaxPackageLength int
	}
	type args struct {
		rev []byte
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int
		want1  int
	}{
		{
			name: "Package Less1",
			fields: fields{
				MaxPackageLength: 8,
			},
			args: args{
				[]byte{1,3},
			},
			want:0,
			want1:0,
		},{
			name: "Package error",
			fields: fields{
				MaxPackageLength: 5,
			},
			args: args{
				[]byte{0,0,0,35,84,64,55},
			},
			want:0,
			want1:2,
		},{
			name: "Package less2",
			fields: fields{
				MaxPackageLength: 5,
			},
			args: args{
				[]byte{0,0,0,5},
			},
			want:0,
			want1:0,
		},{
			name: "Package full",
			fields: fields{
				MaxPackageLength: 1000000,
			},
			args: args{
				[]byte{0,0,0,8,2,0,0,1,1},
			},
			want:8,
			want1:1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &TarsProtocol{}
			maxPackageLength = tt.fields.MaxPackageLength
			got, got1 := p.ParsePackage(tt.args.rev)
			if got != tt.want {
				t.Errorf("ParsePackage() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("ParsePackage() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
func TestTarsProtocol_RequestPack(t *testing.T) {
	req := &requestf.RequestPacket{
		IVersion: basef.TARSVERSION,
		CPacketType: 1,
		IMessageType: 2,
		IRequestId: 3,
		SServantName: "unittest",
		SFuncName: "RequestPack",
		SBuffer: []int8{1,2,3,4,5,6,7,8},
		ITimeout: 4,
		Context: map[string]string{"hello":"tars"},
		Status: map[string]string{"hello":"tarser"},
	}
	p := &TarsProtocol{}
	pack, err := p.RequestPack(req)
	if err != nil {
		t.Errorf("convert RequestPack failed. got err:%v\n", err)
	}
	got := &requestf.RequestPacket{}
	err = got.ReadFrom(codec.NewReader(pack))
	if err != nil {
		t.Errorf("convert RequestPack failed. got err:%v\n", err)
	}

	assert.Equal(t, got, req, "failed to convert RequestPack")
}

func TestTarsProtocol_ResponseUnpack(t *testing.T) {
	resp := &requestf.ResponsePacket{
		IVersion:     basef.TARSVERSION,
		CPacketType:  1,
		IRequestId:   2,
		IMessageType: 3,
		IRet:         4,
		SBuffer:      []int8{1,2,3,4,5,6,7,8},
		Status:       map[string]string{"hello":"tarser"},
		SResultDesc:  "succ",
		Context:      map[string]string{"hello":"tars"},
	}
	buf := codec.NewBuffer()
	err := resp.WriteTo(buf)

	// append package length to package header.
	sbuf := bytes.NewBuffer(nil)
	sbuf.Write(make([]byte, 4))
	sbuf.Write(buf.ToBytes())
	binary.BigEndian.PutUint32(sbuf.Bytes(), uint32(sbuf.Len()))
	if err != nil {
		t.Errorf("write buffer failed. got error:%v\n", err)
	}
	p := &TarsProtocol{}

	got, err := p.ResponseUnpack(sbuf.Bytes())
	if err != nil {
		t.Errorf("unpack response failed. got error:%v\n", err)
	}

	assert.Equal(t, got, resp, "Failed to test ResponseUnpack")

}