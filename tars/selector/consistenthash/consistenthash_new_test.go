package consistenthash

import "testing"

func TestHash(t *testing.T) {
	k := KetamaHashAlg{}
	t.Log(k.Hash("1.1.1.1"))
	d := DefaultHashAlg{}
	t.Log(d.Hash("1.1.1.1"))
}
