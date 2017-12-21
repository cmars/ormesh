// Copyright © 2017 Casey Marshall
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"log"

	"github.com/spf13/cobra"

	"github.com/cmars/ormesh/runner"
)

// exportAddCmd represents the exportAdd command
var exportAddCmd = &cobra.Command{
	Use:   "add [bind addr:]port",
	Short: "Add a service export",
	Long: `Add a service to export as a hidden service. Bind address defaults to 127.0.0.1
if not specified.`,
	Args: cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		err := runner.Run(&runner.ExportAdd{Base: runner.Base{ConfigFile: configFile}}, args)
		if err != nil {
			log.Fatalf("%v", err)
		}
	},
}

func init() {
	exportCmd.AddCommand(exportAddCmd)
}
