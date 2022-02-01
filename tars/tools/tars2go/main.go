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

func (t *importPath) String() string {
	return strings.Join(*t, ":")
}

func (t *importPath) Set(value string) error {
	*t = append(*t, value)
	return nil
}

var (
	gImports  importPath
	gTarsPath string
	gOutdir   string
	gModule   string
	gInclude  string
	includes  []string

	withoutTrace bool
)

func printhelp() {
	bin := os.Args[0]
	if i := strings.LastIndex(bin, "/"); i != -1 {
		bin = bin[i+1:]
	}
	fmt.Printf("Usage: %s [flags] *.tars\n", bin)
	fmt.Printf("       %s -I tars/protocol/res/endpoint [-I ...] QueryF.tars\n", bin)
	fmt.Printf("       %s -include=\"dir1;dir2;dir3\"\n", bin)
	flag.PrintDefaults()
}

func main() {
	flag.Usage = printhelp
	flag.Var(&gImports, "I", "Specify a specific import path")
	flag.StringVar(&gTarsPath, "tarsPath", "github.com/TarsCloud/TarsGo/tars", "Specify the tars source path.")
	flag.StringVar(&gOutdir, "outdir", "", "which dir to put generated code")
	flag.StringVar(&gModule, "module", "", "current go module path")
	flag.StringVar(&gInclude, "include", "", "set search path of tars protocol")
	flag.BoolVar(&withoutTrace, "without-trace", false, "不需要调用链追踪逻辑")
	flag.Parse()

	if flag.NArg() == 0 {
		printhelp()
		os.Exit(0)
	}
	includes = strings.FieldsFunc(gInclude, func(r rune) bool {
		return r == ';' || r == ',' || r == ':' || r == ' '
	})

	for _, filename := range flag.Args() {
		gen := NewGenGo(filename, gModule, gOutdir)
		gen.I = gImports
		gen.tarsPath = gTarsPath
		gen.Gen()
	}
}
