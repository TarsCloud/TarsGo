package endpoint

import (
	"flag"
	"strings"
)

const (
	MAX_WEIGHT int = 100
)

// Parse pares string to struct Endpoint, like tcp -h 10.219.139.142 -p 19386 -t 60000 -v 1 -w 100
func Parse(endpoint string) Endpoint {
	//tcp -h 10.219.139.142 -p 19386 -t 60000 -v 1 -w 100
	proto := endpoint[0:3]
	pFlag := flag.NewFlagSet(proto, flag.ContinueOnError)
	var host, bind string
	var port, timeout, weighttype, weight int

	pFlag.StringVar(&host, "h", "", "host")
	pFlag.IntVar(&port, "p", 0, "port")
	pFlag.IntVar(&timeout, "t", 3000, "timeout")
	pFlag.IntVar(&weighttype, "v", 0, "weighttype")
	pFlag.IntVar(&weight, "w", 0, "weight")
	pFlag.StringVar(&bind, "b", "", "bind")
	pFlag.Parse(strings.Fields(endpoint)[1:])
	istcp := int32(0)
	if proto == "tcp" {
		istcp = int32(1)
	}
	if weighttype == 0 {
		weight = -1
	} else {
		if weight == -1 {
			weight = 100
		}

		if weight > MAX_WEIGHT {
			weight = MAX_WEIGHT
		}
	}

	e := Endpoint{
		Host:       host,
		Port:       int32(port),
		Timeout:    int32(timeout),
		Istcp:      istcp,
		Proto:      proto,
		Bind:       bind,
		WeightType: int32(weighttype),
		Weight:     int32(weight),
	}
	e.Key = e.String()
	return e
}
