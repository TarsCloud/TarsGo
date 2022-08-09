package endpoint

import "fmt"

const (
	UDP int32 = 0
	TCP int32 = 1
	SSL int32 = 2
)

type AuthType int32
type WeightType int32

const (
	AuthTypeNone AuthType = iota
	AuthTypeLocal
)
const (
	ELoop WeightType = iota
	EStaticWeight
)

// Endpoint struct is used record a remote server instance.
type Endpoint struct {
	Host       string
	Port       int32
	Timeout    int32
	Istcp      int32 //need remove
	Grid       int32
	Qos        int32
	Weight     int32
	WeightType int32
	AuthType   int32
	Proto      string
	Bind       string
	Container  string
	SetId      string
	Key        string
}

// String returns readable string for Endpoint
func (e Endpoint) String() string {
	return fmt.Sprintf("%s -h %s -p %d -t %d", e.Proto, e.Host, e.Port, e.Timeout)
}

func (e Endpoint) HashKey() string {
	return e.Host
}

func (e Endpoint) IsTcp() bool {
	return e.Istcp == TCP || e.Istcp == SSL
}

func (e Endpoint) IsUdp() bool {
	return e.Istcp == UDP
}

func (e Endpoint) IsSSL() bool {
	return e.Istcp == SSL
}
