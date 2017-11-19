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
	"net"
	"strconv"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/cmars/ormesh/config"
)

// importAddCmd represents the importAdd command
var importAddCmd = &cobra.Command{
	Use:   "add <remote name> <remote port> <local bind addr>:<local port>",
	Short: "Add a service import",
	Args:  cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		withConfigForUpdate(func(cfg *config.Config) error {
			remoteName, remotePort, localAddr := args[0], args[1], args[2]
			if !IsValidRemoteName(remoteName) {
				return errors.Errorf("invalid remote name %q", remoteName)
			}
			remotePortNum, err := strconv.Atoi(remotePort)
			if err != nil {
				return errors.Errorf("invalid remote port %q", remotePort)
			}
			localAddr, err = NormalizeAddrPort(localAddr)
			if err != nil {
				return errors.Errorf("invalid local address %q", localAddr)
			}
			localHost, localPort, err := net.SplitHostPort(localAddr)
			if err != nil {
				return errors.Errorf("invalid local address %q: %v", localAddr, err)
			}
			localPortNum, err := strconv.Atoi(localPort)
			if err != nil {
				return errors.Errorf("invalid local port %q", localPort)
			}
			newImport := config.Import{
				LocalAddr:  localHost,
				LocalPort:  localPortNum,
				RemotePort: remotePortNum,
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
			for i := range cfg.Node.Remotes[remoteIndex].Imports {
				if cfg.Node.Remotes[i].Imports[i] == newImport {
					return nil
				}
			}
			cfg.Node.Remotes[remoteIndex].Imports = append(
				cfg.Node.Remotes[remoteIndex].Imports,
				newImport)
			return nil
		})
	},
}

func init() {
	importCmd.AddCommand(importAddCmd)
}
