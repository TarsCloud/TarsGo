//
// tars2go
//

package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

type importPath []string

var gTarsPath string

func (t *importPath) String() string {
	return strings.Join(*t, ":")
}

func (t *importPath) Set(value string) error {
	*t = append(*t, value)
	return nil
}

var gOutdir = flag.String("outdir", "", "which dir to put generated code")
var gImports importPath

func printhelp() {
	bin := os.Args[0]
	if i := strings.LastIndex(bin, "/"); i != -1 {
		bin = bin[i+1:]
	}
	fmt.Printf("Usage: %s [flags] *.tars\n", bin)
	fmt.Printf("       %s -I tars/protocol/res/endpoint [-I ...] QueryF.tars\n", bin)
	flag.PrintDefaults()
}

func main() {
	flag.Usage = printhelp
	flag.Var(&gImports, "I", "Specify a specific import path")
	flag.StringVar(&gTarsPath, "tarsPath", "github.com/TarsCloud/TarsGo/tars", "Specify the tars source path.")
	flag.Parse()
	if flag.NArg() == 0 {
		printhelp()
		os.Exit(0)
	}
	for _, filename := range flag.Args() {
		gen := NewGenGo(filename, *gOutdir)
		gen.I = gImports
		gen.tarsPath = gTarsPath
		gen.Gen()
	}
}
