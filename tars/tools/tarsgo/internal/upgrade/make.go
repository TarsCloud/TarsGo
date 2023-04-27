package upgrade

import (
	"fmt"
	"os"
	"path"

	"github.com/TarsCloud/TarsGo/tars/tools/tarsgo/internal/base"
	"github.com/TarsCloud/TarsGo/tars/tools/tarsgo/internal/bindata"
	"github.com/TarsCloud/TarsGo/tars/tools/tarsgo/internal/consts"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// MakeCmd represents the new command.
var MakeCmd = &cobra.Command{
	Use:   "make",
	Short: "Auto upgrade makefile",
	Long: `Auto upgrade makefile. Example: 
tarsgo upgrade make`,
	Run: func(cmd *cobra.Command, args []string) {
		wd, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		makefile := path.Join(wd, "makefile")
		err = base.CopyFile(makefile, makefile, []string{
			`
libpath=${subst :, ,$(GOPATH)}`, "",
			"$(foreach path,$(libpath),$(eval -include $(path)/src/github.com/TarsCloud/TarsGo/tars/makefile.tars))", "-include " + consts.IncludeMakefile,
			"$(foreach path,$(libpath),$(eval -include $(path)/src/github.com/TarsCloud/TarsGo/tars/makefile.tars.gomod))", "-include " + consts.IncludeMakefile,
		})
		if err != nil {
			panic(err)
		}
		err = bindata.RestoreAsset(wd, consts.IncludeMakefile)
		if err != nil {
			panic(err)
		}
		fmt.Println(color.GreenString("upgrade success"))
	},
}
