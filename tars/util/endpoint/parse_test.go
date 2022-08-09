package endpoint

import (
	"fmt"
	"testing"
)

// TestParse tests parsing the endpoint.
func TestParse(t *testing.T) {
	tests := []string{
		"tcp -h 127.0.0.1 -p 19386 -t 60000",
		"udp -h 127.0.0.1 -p 19386 -t 60000",
		"ssl -h 127.0.0.1 -p 19386 -t 60000",
		"ssl -h 127.0.0.1 -p 19386 -t 60000 -g 10 -q 10 -w 10 -v 1 -e 0",
	}
	for _, tt := range tests {
		e2 := Parse(tt)
		fmt.Printf("Parse: %+v\n", e2)
		tars := Endpoint2tars(e2)
		fmt.Printf("Endpoint2tars: %+v\n", tars)
		fmt.Printf("Tars2endpoint: %+v\n", Tars2endpoint(tars))
	}
	fmt.Println(AuthTypeNone, AuthTypeLocal, ELoop, EStaticWeight)
}
