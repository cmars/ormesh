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

// remoteAddCmd represents the remoteAdd command
var remoteAddCmd = &cobra.Command{
	Use:   "add <remote name> <onion address> <client token>",
	Short: "Add a service remote",
	Long: `Add a service remote. The onion address and client token are the values that
were displayed on the remote with the command 'ormesh client add'.`,
	Args: cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		err := runner.Run(&runner.RemoteAdd{Base: runner.Base{ConfigFile: configFile}}, args)
		if err != nil {
			log.Fatalf("%v", err)
		}
	},
}

func init() {
	remoteCmd.AddCommand(remoteAddCmd)
}
