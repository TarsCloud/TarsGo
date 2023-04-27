package roundrobin

import (
	"testing"

	"github.com/TarsCloud/TarsGo/tars/util/endpoint"
)

func TestRoundRobin(t *testing.T) {
	w := New(true)
	w.Add(endpoint.Parse("tcp -h 127.0.0.1 -p 19386 -t 60000 -v 1 -w 3"))
	w.Add(endpoint.Parse("tcp -h 127.0.0.2 -p 19387 -t 60000 -v 1 -w 3"))
	w.Add(endpoint.Parse("tcp -h 127.0.0.3 -p 19388 -t 60000 -v 1 -w 3"))
	w.Add(endpoint.Parse("tcp -h 127.0.0.4 -p 19388 -t 60000 -v 1 -w 3"))

	stats := map[string]int{}
	for i := 0; i < 80; i++ {
		point, _ := w.Select(nil)
		stats[point.Host]++
	}
	t.Log(stats)

	w.Remove(endpoint.Parse("tcp -h 127.0.0.4 -p 19388 -t 60000 -v 1 -w 3"))

	stats = map[string]int{}
	for i := 0; i < 81; i++ {
		point, _ := w.Select(nil)
		stats[point.Host]++
	}
	t.Log(stats)
}
