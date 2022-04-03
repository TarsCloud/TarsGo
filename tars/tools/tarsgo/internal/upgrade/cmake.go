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

// CmakeCmd represents the new command.
var CmakeCmd = &cobra.Command{
	Use:   "cmake",
	Short: "Auto upgrade CMakeLists.txt",
	Long: `Auto upgrade CMakeLists.txt. Example:
tarsgo upgrade cmake`,
	Run: func(cmd *cobra.Command, args []string) {
		wd, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		cmakeListsTxt := path.Join(wd, "CMakeLists.txt")
		err = base.CopyFile(cmakeListsTxt, cmakeListsTxt, []string{
			"${GOPATH}/src/github.com/TarsCloud/TarsGo/", "${CMAKE_CURRENT_SOURCE_DIR}/",
		})
		if err != nil {
			panic(err)
		}
		err = bindata.RestoreAssets(wd, "cmake")
		if err != nil {
			panic(err)
		}
		fmt.Println(color.GreenString("upgrade success"))
	},
}
