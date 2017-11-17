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

// remoteDeleteCmd represents the remoteDelete command
var remoteDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		withConfigForUpdate(func(cfg *config.Config) error {
			remoteName := args[0]
			if !IsValidRemoteName(remoteName) {
				return errors.Errorf("invalid remote %q", remoteName)
			}
			index := -1
			for i := range cfg.Node.Remotes {
				if cfg.Node.Remotes[i].Name == remoteName {
					index = i
				}
			}
			if index < 0 {
				return errors.Errorf("no such remote %q", remoteName)
			}
			cfg.Node.Remotes = append(cfg.Node.Remotes[:index], cfg.Node.Remotes[index+1:]...)
			return nil
		})
	},
}

func init() {
	remoteCmd.AddCommand(remoteDeleteCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// remoteDeleteCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// remoteDeleteCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
