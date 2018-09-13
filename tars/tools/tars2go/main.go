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

var g_tarsPath string

func (self *importPath) String() string {
	return strings.Join(*self, ":")
}

func (self *importPath) Set(value string) error {
	*self = append(*self, value)
	return nil
}

var g_outdir = flag.String("outdir", "", "生成的代码放到哪个目录")
var g_imports importPath

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
	flag.Var(&g_imports, "I", "指定具体的import路径")
	flag.StringVar(&g_tarsPath, "tarsPath", "github.com/TarsCloud/TarsGo/tars", "specify the tars source path.")
	flag.Parse()
	if flag.NArg() == 0 {
		printhelp()
		os.Exit(0)
	}
	for _, filename := range flag.Args() {
		gen := NewGenGo(filename, *g_outdir)
		gen.I = g_imports
		gen.tarsPath = g_tarsPath
		gen.Gen()
	}
}
