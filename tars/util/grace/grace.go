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
		fd := uint64(f.Fd())
		val = fmt.Sprint(fd)
		os.Setenv(key, val)
	}
	return ln, err
}

func CreateUDPConn(addr string) (*net.UDPConn, error) {
	key := fmt.Sprintf("%s_%s_%s", ListenFdEnvPrefix, "udp", addr)
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
		return conn.(*net.UDPConn), nil
	}
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
		os.Setenv(key, val)
	}
	return conn, err
}

type filer interface {
	File() (*os.File, error)
}
