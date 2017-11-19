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

// exportAddCmd represents the exportAdd command
var exportAddCmd = &cobra.Command{
	Use:   "add [bind addr:]port",
	Short: "Add a service export",
	Long: `Add a service to export as a hidden service. Bind address defaults to 127.0.0.1
if not specified.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		withConfigForUpdate(func(cfg *config.Config) error {
			exportAddr, err := NormalizeAddrPort(args[0])
			if err != nil {
				return errors.Errorf("invalid export address %q", args[0])
			}
			index := -1
			for i := range cfg.Node.Service.Exports {
				if cfg.Node.Service.Exports[i] == exportAddr {
					index = i
					break
				}
			}
			if index < 0 {
				cfg.Node.Service.Exports = append(cfg.Node.Service.Exports, exportAddr)
			}
			return nil
		})
	},
}

func init() {
	exportCmd.AddCommand(exportAddCmd)
}
