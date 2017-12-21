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

var displayQR bool

// clientAddCmd represents the clientAdd command
var clientAddCmd = &cobra.Command{
	Use:   "add <client name>",
	Short: "Add a client authorization",
	Long: `Create an auth token allowing a client to access exported services. The token
should be securely transmitted to the client.`,
	Example: `
  $ ormesh client add my-MacBook
  Y5Cfw7A5RhP8Rd7xGYfD8N4oyEBpBWNR+6Qkgrbepk0=

  Then paste this token as an argument to 'ormesh remote add':

  $ ormesh remote add my-server Y5Cfw7A5RhP8Rd7xGYfD8N4oyEBpBWNR+6Qkgrbepk0=`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		err := runner.Run(&runner.ClientAdd{
			Base:      runner.Base{ConfigFile: configFile},
			DisplayQR: displayQR,
		}, args)
		if err != nil {
			log.Fatalf("%v", err)
		}
	},
}

func init() {
	clientAddCmd.Flags().BoolVarP(&displayQR, "qr", "", false, "Display Orbot client cookie QR code")
	clientCmd.AddCommand(clientAddCmd)
}
