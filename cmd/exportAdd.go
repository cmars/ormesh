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

	"github.com/cmars/ormesh/agent"
	"github.com/cmars/ormesh/config"
)

// exportAddCmd represents the exportAdd command
var exportAddCmd = &cobra.Command{
	Use:   "add",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		withConfigForUpdate(func(cfg *config.Config) {
			exportAddr, err := NormalizeAddrPort(args[0])
			if err != nil {
				log.Fatalf("invalid export address %q", args[0])
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
			} else {
				return
			}
			a, err := agent.New(cfg)
			if err != nil {
				log.Fatalf("failed to initialize agent: %v", err)
			}
			err = a.UpdateServices(&cfg.Node.Service)
			if err != nil {
				log.Fatalf("failed to update tor: %v", err)
			}
		})
	},
}

func init() {
	exportCmd.AddCommand(exportAddCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// exportAddCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// exportAddCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
