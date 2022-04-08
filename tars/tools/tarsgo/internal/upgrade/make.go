package upgrade

import (
	"fmt"
	"github.com/TarsCloud/TarsGo/tars/tools/tarsgo/internal/base"
	"github.com/TarsCloud/TarsGo/tars/tools/tarsgo/internal/bindata"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"os"
	"path"
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
			"$(foreach path,$(libpath),$(eval -include $(path)/src/github.com/TarsCloud/TarsGo/tars/makefile.tars))", "-include scripts/makefile.tars.gomod",
			"$(foreach path,$(libpath),$(eval -include $(path)/src/github.com/TarsCloud/TarsGo/tars/makefile.tars.gomod))", "-include scripts/makefile.tars.gomod",
		})
		if err != nil {
			panic(err)
		}
		err = bindata.RestoreAsset(wd, "scripts/makefile.tars.gomod")
		if err != nil {
			panic(err)
		}
		fmt.Println(color.GreenString("upgrade success"))
	},
}
