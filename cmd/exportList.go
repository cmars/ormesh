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

	"github.com/spf13/cobra"

	"github.com/cmars/ormesh/config"
)

// exportListCmd represents the exportList command
var exportListCmd = &cobra.Command{
	Use:   "list",
	Short: "List service exports",
	Run: func(cmd *cobra.Command, args []string) {
		withConfig(func(cfg *config.Config) error {
			for _, export := range cfg.Node.Service.Exports {
				fmt.Printf("%#v\n", export)
			}
			return nil
		})
	},
}

func init() {
	exportCmd.AddCommand(exportListCmd)
}
