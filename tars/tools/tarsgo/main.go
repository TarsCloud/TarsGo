/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package main

import (
	"github.com/TarsCloud/TarsGo/tars/tools/tarsgo/internal/cmake"
	"github.com/TarsCloud/TarsGo/tars/tools/tarsgo/internal/consts"
	"github.com/TarsCloud/TarsGo/tars/tools/tarsgo/internal/make"
	"github.com/TarsCloud/TarsGo/tars/tools/tarsgo/internal/upgrade"
	"github.com/spf13/cobra"
	"log"
)

var rootCmd = &cobra.Command{
	Use:     "tarsgo",
	Short:   "tarsgo: An elegant toolkit for Go microservices.",
	Long:    `tarsgo: An elegant toolkit for Go microservices.`,
	Version: consts.Release,
}

func init() {
	rootCmd.AddCommand(make.CmdNew)
	rootCmd.AddCommand(cmake.CmdNew)
	rootCmd.AddCommand(upgrade.CmdNew)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
