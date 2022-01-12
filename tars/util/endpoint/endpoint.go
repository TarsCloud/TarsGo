package endpoint

import "fmt"

const (
	UDP int32 = 0
	TCP int32 = 1
	SSL int32 = 2
)

//Endpoint struct is used record a remote server instance.
type Endpoint struct {
	Host      string
	Port      int32
	Timeout   int32
	Istcp     int32 //need remove
	Proto     string
	Bind      string
	Container string
	SetId     string
	Key       string
}

// String returns readable string for Endpoint
func (e Endpoint) String() string {
	return fmt.Sprintf("%s -h %s -p %d -t %d -d %s", e.Proto, e.Host, e.Port, e.Timeout, e.Container)
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
