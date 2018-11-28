package grace

import (
	"fmt"
	"net"
	"os"
	"strconv"
)

var (
	// InheritFdPrefix marks the fd inherited from parent process
	InheritFdPrefix = "LISTEN_FD_INHERIT"
	// ListenFdPrefix marks the fd listened by current process
	ListenFdPrefix = "LISTEN_FD_CURRENT"
)

// CreateListener creates a listener from inherited fd
// if there is no inherited fd, create a now one.
func CreateListener(proto string, addr string) (net.Listener, error) {
	key := fmt.Sprintf("%s_%s_%s", InheritFdPrefix, proto, addr)
	nowKey := fmt.Sprintf("%s_%s_%s", ListenFdPrefix, proto, addr)
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
		os.Setenv(nowKey, val)
		file.Close()
		return ln, nil
	}
	// not inherit, create new
	ln, err := net.Listen(proto, addr)
	if err == nil {
		f, _ := ln.(filer).File()
		val = fmt.Sprint(f.Fd())
		os.Setenv(nowKey, val)
	}
	return ln, err
}

// CreateUDPConn creates a udp connection from inherited fd
// if there is no inherited fd, create a now one.
func CreateUDPConn(addr string) (*net.UDPConn, error) {
	proto := "udp"
	key := fmt.Sprintf("%s_%s_%s", InheritFdPrefix, proto, addr)
	nowKey := fmt.Sprintf("%s_%s_%s", ListenFdPrefix, proto, addr)
	val := os.Getenv(key)
	for val != "" {
		fd, err := strconv.Atoi(val)
		if err != nil {
			break
		}
		file := os.NewFile(uintptr(fd), "listener")
		conn, err := net.FileConn(file)
		if err != nil {
			file.Close()
			break
		}
		os.Setenv(nowKey, val)
		file.Close()
		return conn.(*net.UDPConn), nil
	}
	// not inherit, create new
	uaddr, err := net.ResolveUDPAddr("udp4", addr)
	if err != nil {
		return nil, err
	}
	conn, err := net.ListenUDP("udp4", uaddr)
	if err != nil {
		return nil, err
	}
	if err == nil {
		f, _ := conn.File()
		val = fmt.Sprint(f.Fd())
		os.Setenv(nowKey, val)
	}
	return conn, err
}

type filer interface {
	File() (*os.File, error)
}
