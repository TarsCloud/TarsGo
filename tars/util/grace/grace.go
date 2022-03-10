package grace

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"sync"
)

var (
	// InheritFdPrefix marks the fd inherited from parent process
	InheritFdPrefix = "LISTEN_FD_INHERIT"

	allListenFds *sync.Map
)

func init() {
	allListenFds = &sync.Map{}
}

// CreateListener creates a listener from inherited fd
// if there is no inherited fd, create a now one.
func CreateListener(proto string, addr string) (net.Listener, error) {
	key := fmt.Sprintf("%s_%s_%s", InheritFdPrefix, proto, addr)
	val := os.Getenv(key)
	if val != "" {
		fd, err := strconv.Atoi(val)
		if err != nil {
			goto CreateTcp
		}
		file := os.NewFile(uintptr(fd), "listener")
		ln, err := net.FileListener(file)
		if err != nil {
			file.Close()
			goto CreateTcp
		}
		allListenFds.Store(key, ln)
		return ln, nil
	}
	// not inherit, create new
CreateTcp:
	ln, err := net.Listen(proto, addr)
	if err == nil {
		allListenFds.Store(key, ln)
	}
	return ln, err
}

// CreateUDPConn creates a udp connection from inherited fd
// if there is no inherited fd, create a now one.
func CreateUDPConn(addr string) (*net.UDPConn, error) {
	proto := "udp"
	key := fmt.Sprintf("%s_%s_%s", InheritFdPrefix, proto, addr)
	val := os.Getenv(key)
	if val != "" {
		fd, err := strconv.Atoi(val)
		if err != nil {
			goto CreateUdp
		}
		file := os.NewFile(uintptr(fd), "listener")
		conn, err := net.FileConn(file)
		if err != nil {
			file.Close()
			goto CreateUdp
		}
		udpConn := conn.(*net.UDPConn)
		allListenFds.Store(key, udpConn)
		return udpConn, nil
	}
	// not inherit, create new
CreateUdp:
	uaddr, err := net.ResolveUDPAddr("udp4", addr)
	if err != nil {
		return nil, err
	}
	conn, err := net.ListenUDP("udp4", uaddr)
	if err == nil {
		allListenFds.Store(key, conn)
	}
	return conn, err
}

// GetAllListenFiles returns all listen files
func GetAllListenFiles() map[string]*os.File {
	files := make(map[string]*os.File)
	allListenFds.Range(func(k, v interface{}) bool {
		key := k.(string)
		val := v.(filer)
		if file, err := val.File(); err == nil {
			files[key] = file
		}
		return true
	})
	return files
}

type filer interface {
	File() (*os.File, error)
}
