package upgrade

import (
	"fmt"
	"github.com/TarsCloud/TarsGo/tars/tools/tarsgo/internal/base"
	"github.com/spf13/cobra"
)

// CmdNew UpgradeCmd represents the new command.
var CmdNew = &cobra.Command{
	Use:   "upgrade",
	Short: "Auto upgrade tarsgo and tars2go",
	Long: `Auto upgrade tarsgo and tars2go. Example:
tarsgo upgrade`,
	Run: run,
}

var force bool

func init() {
	CmdNew.Flags().BoolVarP(&force, "force", "f", false, "force upgrade tarsgo and tars2go")
	CmdNew.AddCommand(MakeCmd)
	CmdNew.AddCommand(CmakeCmd)
}

func run(cmd *cobra.Command, args []string) {
	if force {
		err := base.GoInstall(
			"github.com/TarsCloud/TarsGo/tars/tools/tarsgo",
			"github.com/TarsCloud/TarsGo/tars/tools/tars2go",
			"github.com/TarsCloud/TarsGo/tars/tools/pb2tarsgo/protoc-gen-go",
		)
		if err != nil {
			fmt.Println(err)
		}
	}
}
