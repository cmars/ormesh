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
	"strconv"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/cmars/ormesh/config"
)

// importDeleteCmd represents the importDelete command
var importDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a service import",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		withConfigForUpdate(func(cfg *config.Config) error {
			remoteName, remotePort := args[0], args[1]
			if !IsValidRemoteName(remoteName) {
				return errors.Errorf("invalid remote name %q", remoteName)
			}
			remotePortNum, err := strconv.Atoi(remotePort)
			if err != nil {
				return errors.Errorf("invalid remote port %q", remotePort)
			}
			remoteIndex := -1
			for i := range cfg.Node.Remotes {
				if cfg.Node.Remotes[i].Name == remoteName {
					remoteIndex = i
					break
				}
			}
			if remoteIndex < 0 {
				return errors.Errorf("no such remote: %q", remoteName)
			}
			var imports []config.Import
			for i := range cfg.Node.Remotes[remoteIndex].Imports {
				if cfg.Node.Remotes[remoteIndex].Imports[i].RemotePort != remotePortNum {
					imports = append(imports, cfg.Node.Remotes[remoteIndex].Imports[i])
				}
			}
			cfg.Node.Remotes[remoteIndex].Imports = imports
			return nil
		})
	},
}

func init() {
	importCmd.AddCommand(importDeleteCmd)
}
