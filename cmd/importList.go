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
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/cmars/ormesh/config"
)

// importListCmd represents the importList command
var importListCmd = &cobra.Command{
	Use:   "list",
	Short: "List service imports",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		withConfig(func(cfg *config.Config) error {
			remoteName := args[0]
			if !IsValidRemoteName(remoteName) {
				return errors.Errorf("invalid remote name %q", remoteName)
			}
			// TODO: use specified remote
			for _, remote := range cfg.Node.Remotes {
				if remote.Name == remoteName {
					for _, import_ := range remote.Imports {
						fmt.Printf("%#v\n", import_)
					}
					return nil
				}
			}
			return errors.Errorf("no such remote %q", remoteName)
		})
	},
}

func init() {
	importCmd.AddCommand(importListCmd)
}
