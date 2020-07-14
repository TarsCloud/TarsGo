package main

import (
	"testing"
)

func TestGenGo_Gen(t *testing.T) {
	var filename = "/home/yifabao/go/src/fabaoyi.com/QuwanyunApp/PhoneRedPacketServer/ReportInfo.tars"
	gOutdir = "/home/yifabao/go/src/fabaoyi.com/QuwanyunApp/PhoneRedPacketServer/tars-protocol"
	gModule = "fabaoyi.com/QuwanyunApp/PhoneRedPacketServer"
	gen := NewGenGo(filename, gModule, gOutdir)
	gen.I = gImports
	gen.tarsPath = gTarsPath
	gen.Gen()
}
