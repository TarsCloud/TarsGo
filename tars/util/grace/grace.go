package grace

import (
	"fmt"
	"net"
	"os"
	"strconv"
)

var ListenFdEnvPrefix = "LISTEN_FDS"

func CreateListener(proto string, addr string) (net.Listener, error) {
	key := fmt.Sprintf("%s_%s_%s", ListenFdEnvPrefix, proto, addr)
	val := os.Getenv(key)
	for val != "" {
		fd, err := strconv.Atoi(val)
		if err != nil {
			break
		}
		file := os.NewFile(uintptr(fd), "listener")
		ln, err := net.FileListener(file)
		if err != nil {
			file.Close()
			break
		}
		return ln, nil
	}
	ln, err := net.Listen(proto, addr)
	if err == nil {
		f, _ := ln.(filer).File()
		val = fmt.Sprint(f.Fd())
		os.Setenv(key, val)
	}
	return ln, err
}

type filer interface {
	File() (*os.File, error)
}
