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

// exportDeleteCmd represents the exportDelete command
var exportDeleteCmd = &cobra.Command{
	Use:   "delete [bind addr:]port",
	Short: "Delete a service export",
	Long: `Delete a service from the list of exports. Bind address must match the existing
entry to be deleted, defaulting to 127.0.0.1 if not specified.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		withConfigForUpdate(func(cfg *config.Config) error {
			localAddr, err := NormalizeAddrPort(args[0])
			if err != nil {
				return errors.Errorf("invalid local address %q", args[0])
			}
			index := -1
			for i := range cfg.Node.Service.Exports {
				if cfg.Node.Service.Exports[i].LocalAddr == localAddr {
					index = i
					break
				}
			}
			if index == -1 {
				return errors.Errorf("no such export: %q", localAddr)
			}
			cfg.Node.Service.Exports = append(
				cfg.Node.Service.Exports[:index],
				cfg.Node.Service.Exports[index+1:]...)
			return nil
		})
	},
}

func init() {
	exportCmd.AddCommand(exportDeleteCmd)
}
