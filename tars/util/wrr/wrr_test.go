package wrr

import (
	"fmt"
	"testing"

	"github.com/TarsCloud/TarsGo/tars/util/endpoint"
)

// TestWrrBalance tests pasing the wrr.
func TestWrrBalance(t *testing.T) {
	mapValues := make(map[string]int, 0)
	wrr := NewWrrBalance()
	e1 := endpoint.Parse("tcp -h 127.0.0.1 -p 19386 -t 60000 -v 1 -w 10")
	e2 := endpoint.Parse("udp -h 127.0.0.1 -p 19387 -t 60000 -v 1 -w 100")
	e3 := endpoint.Parse("udp -h 127.0.0.1 -p 19388 -t 60000 -v 0 -w 0")

	wrr.Add(e1, e1.Weight)
	wrr.Add(e2, e2.Weight)
	wrr.Add(e3, e3.Weight)

	for i := 0; i < 100; i++ {
		ep, _ := wrr.Next()
		mapValues[ep.String()]++
	}

	for ep, cnt := range mapValues {
		fmt.Println(ep, ":", cnt)
	}
}
