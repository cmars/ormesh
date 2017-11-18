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
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/cmars/ormesh/config"
)

// clientDeleteCmd represents the clientDelete command
var clientDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a client authorization",
	Long: `
Delete a client authorization.

  This command will remove the auth token added for the named client, revoking
  access permanently.

Usage:

	$ ormesh client delete <client name>

Arguments:

  'client name' is a locally unique name that identifies the client for the
  purpose of managing its authorization.
`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		withConfigForUpdate(func(cfg *config.Config) error {
			clientName := args[0]
			if !IsValidClientName(clientName) {
				return errors.Errorf("invalid client name %q", clientName)
			}
			index := -1
			for i := range cfg.Node.Service.Clients {
				if cfg.Node.Service.Clients[i].Name == clientName {
					index = i
					break
				}
			}
			if index == -1 {
				return errors.Errorf("no such client %q", clientName)
			}
			cfg.Node.Service.Clients = append(
				cfg.Node.Service.Clients[:index],
				cfg.Node.Service.Clients[index+1:]...)
			return nil
		})
	},
}

func init() {
	clientCmd.AddCommand(clientDeleteCmd)
}
