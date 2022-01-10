package tools

import "net"

var cacheIP string

func init() {
	GetLocalIP()
}

func GetLocalIP() string {
	if cacheIP != "" {
		return cacheIP
	}
	conn, err := net.Dial("udp", "10.0.0.1:80")
	if err != nil {
		return "0.0.0.0"
	}
	defer conn.Close()

	addr := conn.LocalAddr().(*net.UDPAddr)
	if addr == nil {
		cacheIP = "0.0.0.0"
	} else {
		cacheIP = addr.IP.String()
	}
	return cacheIP
}
