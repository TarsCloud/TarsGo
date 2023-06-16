//
// tars2go
//

package main

import (
	"flag"
	"os"

	"github.com/TarsCloud/TarsGo/tars/tools/tars2go/gencode"
	"github.com/TarsCloud/TarsGo/tars/tools/tars2go/options"
)

func main() {
	opt := options.NewOptions()
	if flag.NArg() == 0 {
		opt.PrintHelp()
		os.Exit(0)
	}

	for _, filename := range flag.Args() {
		gen := gencode.NewGenGo(opt, filename)
		gen.Gen()
	}
}
