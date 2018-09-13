package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"strconv"
)

func hello(conn net.Conn, name string) {
	payload := []byte(name)
	pkg := make([]byte, 4+len(payload))
	binary.BigEndian.PutUint32(pkg[:4], uint32(len(pkg)))
	copy(pkg[4:], payload)
	conn.Write(pkg)
	buf := make([]byte, 1024*4)
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(buf[4:n]))
}

func main() {
	name := "Bob"
	if len(os.Args) == 2 {
		name = os.Args[1]
	}
	addr, err := net.ResolveUDPAddr("udp", "localhost:3333")
	if err != nil {
		fmt.Println("Can't resolve address: ", err)
		os.Exit(1)
	}
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		fmt.Println("Can't dial: ", err)
		os.Exit(1)
	}

	defer conn.Close()
	for i := 0; i < 5; i++ {
		hello(conn, name+strconv.Itoa(i))
	}

	os.Exit(0)
}
