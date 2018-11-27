package grace

import (
	"fmt"
	"os"
	"testing"
)

func TestListener(t *testing.T) {
	proto, addr := "tcp", "localhost:3333"
	_, err := CreateListener(proto, addr)
	key := fmt.Sprintf("%s_%s_%s", ListenFdEnvPrefix, proto, addr)
	fmt.Println(err, key, os.Getenv(key))

	addr = "localhost:3334"
	_, err = CreateUDPConn(addr)
	key = fmt.Sprintf("%s_%s_%s", ListenFdEnvPrefix, "udp", addr)
	fmt.Println(err, key, os.Getenv(key))
}
