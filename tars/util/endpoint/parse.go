package endpoint

import (
	"flag"
	"strings"
)

// Parse pares string to struct Endpoint, like tcp -h 10.219.139.142 -p 19386 -t 60000
func Parse(endpoint string) Endpoint {
	// tcp -h 10.219.139.142 -p 19386 -t 60000
	proto := endpoint[0:3]
	pFlag := flag.NewFlagSet(proto, flag.ContinueOnError)
	var host, bind string
	var port, timeout, grid, qos, weight, weightType, authType int
	pFlag.StringVar(&host, "h", "", "host")
	pFlag.IntVar(&port, "p", 0, "port")
	pFlag.IntVar(&timeout, "t", 3000, "timeout")
	pFlag.IntVar(&grid, "g", 0, "grid")
	pFlag.IntVar(&qos, "q", 0, "qos")
	pFlag.IntVar(&weight, "w", -1, "weight")
	pFlag.IntVar(&weightType, "v", 0, "weight type") // 权重类型
	pFlag.IntVar(&authType, "e", 0, "auth type")     // 鉴权类型: enum AUTH_TYPE { AUTH_TYPENONE = 0, AUTH_TYPELOCAL = 1};
	pFlag.StringVar(&bind, "b", "", "bind")
	_ = pFlag.Parse(strings.Fields(endpoint)[1:])
	isTcp := int32(0)
	if proto == "tcp" {
		isTcp = int32(1)
	} else if proto == "ssl" {
		proto = "tcp"
		isTcp = int32(2)
	}
	if weightType != 0 && (weight == -1 || weight > 100) {
		weight = 100
	}
	e := Endpoint{
		Host:       host,
		Port:       int32(port),
		Timeout:    int32(timeout),
		Istcp:      isTcp,
		Grid:       int32(grid),
		Qos:        int32(qos),
		Weight:     int32(weight),
		WeightType: int32(weightType),
		AuthType:   int32(AuthType(authType)),
		Proto:      proto,
		Bind:       bind,
	}
	e.Key = e.String()
	return e
}
