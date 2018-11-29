package grace

import (
	"fmt"
	"os"
	"os/exec"
	"testing"
)

func TestListener(t *testing.T) {
	proto, addr := "tcp", "localhost:3333"
	_, err := CreateListener(proto, addr)
	key := fmt.Sprintf("%s_%s_%s", ListenFdPrefix, proto, addr)
	fmt.Println(err, key, os.Getenv(key))
	cmd := exec.Command("lsof", "-p", fmt.Sprint(os.Getpid()))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Start()
	cmd.Wait()
	addr = "localhost:3334"
	_, err = CreateUDPConn(addr)
	key = fmt.Sprintf("%s_%s_%s", ListenFdPrefix, "udp", addr)
	fmt.Println(err, key, os.Getenv(key))
	cmd = exec.Command("lsof", "-p", fmt.Sprint(os.Getpid()))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Start()
	cmd.Wait()
}
