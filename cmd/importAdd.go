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
	"net"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/cmars/ormesh/config"
)

// importAddCmd represents the importAdd command
var importAddCmd = &cobra.Command{
	Use:   "add",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Args: cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		remoteName, remotePort, localAddr := args[0], args[1], args[2]
		if !IsValidRemoteName(remoteName) {
			log.Fatalf("invalid remote name %q", remoteName)
		}
		remotePortNum, err := strconv.Atoi(remotePort)
		if err != nil {
			log.Fatalf("invalid remote port %q", remotePort)
		}
		localAddr, err = NormalizeAddrPort(localAddr)
		if err != nil {
			log.Fatalf("invalid local address %q", localAddr)
		}
		localHost, localPort, err := net.SplitHostPort(localAddr)
		if err != nil {
			log.Fatalf("invalid local address %q: %v", localAddr, err)
		}
		localPortNum, err := strconv.Atoi(localPort)
		if err != nil {
			log.Fatalf("invalid local port %q", localPort)
		}
		newImport := config.Import{
			LocalAddr:  localHost,
			LocalPort:  localPortNum,
			RemotePort: remotePortNum,
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
			for i := range cfg.Node.Remotes[remoteIndex].Imports {
				if cfg.Node.Remotes[i].Imports[i] == newImport {
					return
				}
			}
			cfg.Node.Remotes[remoteIndex].Imports = append(
				cfg.Node.Remotes[remoteIndex].Imports,
				newImport)
		})
	},
}

func init() {
	importCmd.AddCommand(importAddCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// importAddCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// importAddCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
