package base

import (
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
)

func GetArgs(cmd *cobra.Command, args []string) (app, server, servant, goModuleName string, err error) {
	if len(args) == 0 {
		prompt := &survey.Input{
			Message: "What is project app ?",
			Help:    "Created project app.",
		}
		err = survey.AskOne(prompt, &app)
		if err != nil || app == "" {
			return
		}
		prompt = &survey.Input{
			Message: "What is project server ?",
			Help:    "Created project server.",
		}
		err = survey.AskOne(prompt, &server)
		if err != nil || server == "" {
			return
		}
		prompt = &survey.Input{
			Message: "What is project servant ?",
			Help:    "Created project servant.",
		}
		err = survey.AskOne(prompt, &servant)
		if err != nil || servant == "" {
			return
		}
		prompt = &survey.Input{
			Message: "What is project GoModuleName ?",
			Help:    "Created project GoModuleName.",
		}
		err = survey.AskOne(prompt, &goModuleName)
		if err != nil || goModuleName == "" {
			return
		}
	} else if len(args) != 4 {
		_ = cmd.Help()
		err = fmt.Errorf("args: %+v", args)
		return
	} else {
		app = args[0]
		server = args[1]
		servant = args[2]
		goModuleName = args[3]
	}
	return app, server, servant, goModuleName, nil
}
