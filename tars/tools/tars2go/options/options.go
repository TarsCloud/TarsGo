package options

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

type ImportPath []string

type Options struct {
	Imports  ImportPath
	TarsPath string
	Outdir   string
	Module   string
	Include  string
	Includes []string

	WithoutTrace bool
	// gen
	E                bool
	AddServant       bool
	ModuleCycle      bool
	ModuleUpper      bool
	JsonOmitEmpty    bool
	DispatchReporter bool
	Debug            bool
}

func NewOptions() *Options {
	o := &Options{}
	o.initFlags()
	return o
}

func (o *Options) initFlags() {
	flag.Usage = o.PrintHelp
	flag.Var(&o.Imports, "I", "Specify a specific import path")
	flag.StringVar(&o.TarsPath, "tarsPath", "github.com/TarsCloud/TarsGo/tars", "Specify the tars source path.")
	flag.StringVar(&o.Outdir, "outdir", "", "which dir to put generated code")
	flag.StringVar(&o.Module, "module", "", "current go module path")
	flag.StringVar(&o.Include, "include", "", "set search path of tars protocol")
	flag.BoolVar(&o.WithoutTrace, "without-trace", false, "no call chain tracking logic required")

	// gen options
	flag.BoolVar(&o.E, "E", false, "Generate code before fmt for troubleshooting")
	flag.BoolVar(&o.AddServant, "add-servant", true, "Generate AddServant function")
	flag.BoolVar(&o.ModuleCycle, "module-cycle", false, "support jce module cycle include(do not support jce file cycle include)")
	flag.BoolVar(&o.ModuleUpper, "module-upper", false, "native module names are supported, otherwise the system will upper the first letter of the module name")
	flag.BoolVar(&o.JsonOmitEmpty, "json-omitempty", false, "Generate json omitempty support")
	flag.BoolVar(&o.DispatchReporter, "dispatch-reporter", false, "Dispatch reporter support")
	flag.BoolVar(&o.Debug, "debug", false, "enable debug mode")
	flag.Parse()

	o.Includes = strings.FieldsFunc(o.Include, func(r rune) bool {
		return r == ';' || r == ',' || r == ':' || r == ' '
	})
}

func (o *Options) PrintHelp() {
	bin := os.Args[0]
	if i := strings.LastIndex(bin, "/"); i != -1 {
		bin = bin[i+1:]
	}
	fmt.Printf("Usage: %s [flags] *.tars\n", bin)
	fmt.Printf("       %s -I tars/protocol/res/endpoint [-I ...] QueryF.tars\n", bin)
	fmt.Printf("       %s -include=\"dir1;dir2;dir3\"\n", bin)
	flag.PrintDefaults()
}

func (ip *ImportPath) String() string {
	return strings.Join(*ip, ":")
}

func (ip *ImportPath) Set(value string) error {
	*ip = append(*ip, value)
	return nil
}
