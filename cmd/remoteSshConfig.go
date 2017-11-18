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
	"net"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/cmars/ormesh/config"
)

// remoteSshConfigCmd represents the sshConfig command
var remoteSshConfigCmd = &cobra.Command{
	Use:   "ssh-config",
	Short: "Print ssh-config(5) stanza for a remote",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		withConfig(func(cfg *config.Config) error {
			remoteName := args[0]
			if !IsValidRemoteName(remoteName) {
				return errors.Errorf("invalid remote %q", remoteName)
			}
			_, socksPort, err := net.SplitHostPort(cfg.Node.Agent.SocksAddr)
			if err != nil {
				return errors.Errorf("invalid SocksAddr %q", cfg.Node.Agent.SocksAddr)
			}
			for _, remote := range cfg.Node.Remotes {
				if remote.Name == remoteName {
					fmt.Printf(`Host %s
  ProxyCommand nc -X 5 -x 127.0.0.1:%d %%h %%p
  Hostname %s
`, remoteName, socksPort, remote.Address)
					return nil
				}
			}
			return errors.Errorf("no such remote %q", remoteName)
		})
	},
}

func init() {
	remoteCmd.AddCommand(remoteSshConfigCmd)
}
