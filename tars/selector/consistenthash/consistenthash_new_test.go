package consistenthash

import (
	"testing"

	"github.com/TarsCloud/TarsGo/tars/util/endpoint"
)

func TestHash(t *testing.T) {
	k := KetamaHashAlg{}
	t.Log(k.Hash("1.1.1.1"))
	d := DefaultHashAlg{}
	t.Log(d.Hash("1.1.1.1"))
}

func TestConsistentHash(t *testing.T) {
	ch := New(false, DefaultHash)
	ch.Add(endpoint.Endpoint{
		Host: "10.160.129.102",
		Qos:  2,
	})
	ch.Add(endpoint.Endpoint{
		Host: "10.160.129.105",
		Qos:  1,
	})
	ch.printNode()
	ep, _ := ch.FindInt32(hashFn("#12723353"))
	t.Log("#12723353", hashFn("#12723353"), ep.Host)
	ep, _ = ch.FindInt32(hashFn("#12723353_native"))
	t.Log("#12723353_native", hashFn("#12723353_native"), ep.Host)
}

func TestKetamaHashAlg_Hash(t *testing.T) {
	tests := []struct {
		name string
		str  string
		want uint32
	}{
		{
			name: "1.1.1.1",
			str:  "1.1.1.1",
			want: 329942752,
		},
		{
			name: "2.2.2.2",
			str:  "2.2.2.2",
			want: 329942752,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := KetamaHashAlg{}
			if code := k.Hash(tt.str); code != tt.want {
				t.Errorf("Hash() = %v, want %v", code, tt.want)
			}
		})
	}
}

func hashFn(str string) uint32 {
	var value uint32
	for _, c := range str {
		value += uint32(c)
		value += value << 10
		value ^= value >> 6
	}
	value += value << 3
	value ^= value >> 11
	value += value << 15
	if value == 0 {
		return 1
	}
	return value
}
