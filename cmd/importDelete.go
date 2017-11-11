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
	"strconv"

	"github.com/spf13/cobra"

	"github.com/cmars/ormesh/config"
)

// importDeleteCmd represents the importDelete command
var importDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		remoteName, remotePort := args[0], args[1]
		if !IsValidRemoteName(remoteName) {
			log.Fatalf("invalid remote name %q", remoteName)
		}
		remotePortNum, err := strconv.Atoi(remotePort)
		if err != nil {
			log.Fatalf("invalid remote port %q", remotePort)
		}
		withConfigForUpdate(func(cfg *config.Config) {
			remoteIndex := -1
			for i := range cfg.Node.Remotes {
				if cfg.Node.Remotes[i].Name == remoteName {
					remoteIndex = i
					break
				}
			}
			if remoteIndex < 0 {
				log.Fatalf("no such remote: %q", remoteName)
			}
			var imports []config.Import
			for i := range cfg.Node.Remotes[remoteIndex].Imports {
				if cfg.Node.Remotes[remoteIndex].Imports[i].RemotePort != remotePortNum {
					imports = append(imports, cfg.Node.Remotes[remoteIndex].Imports[i])
				}
			}
			cfg.Node.Remotes[remoteIndex].Imports = imports
		})
	},
}

func init() {
	importCmd.AddCommand(importDeleteCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// importDeleteCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// importDeleteCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
