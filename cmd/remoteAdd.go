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

// remoteAddCmd represents the remoteAdd command
var remoteAddCmd = &cobra.Command{
	Use:   "add <remote name> <client token>",
	Short: "Add a service remote",
	Long: `Add a service remote. The client token is the value that was displayed on the
remote with the command 'ormesh client add'.`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		withConfigForUpdate(func(cfg *config.Config) error {
			remoteName, token := args[0], args[1]
			if !IsValidRemoteName(remoteName) {
				return errors.Errorf("invalid remote name %q", remoteName)
			}
			clientAddr, clientAuth, err := ParseClientToken(token)
			if err != nil {
				return errors.Errorf("invalid client token: %v", err)
			}
			for i := range cfg.Node.Remotes {
				if cfg.Node.Remotes[i].Name == remoteName {
					return errors.Errorf("remote %q already exists", remoteName)
				}
			}
			remote := config.Remote{
				Name:    remoteName,
				Address: clientAddr,
				Auth:    clientAuth,
			}
			cfg.Node.Remotes = append(cfg.Node.Remotes, remote)
			return nil
		})
	},
}

func init() {
	remoteCmd.AddCommand(remoteAddCmd)
}
