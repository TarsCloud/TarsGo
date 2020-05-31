package endpoint

import "fmt"

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
