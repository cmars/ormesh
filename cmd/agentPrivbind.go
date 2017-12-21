// +build linux

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

	"github.com/cmars/ormesh/runner"
)

// agentPrivbindCmd represents the agentPrivbind command
var agentPrivbindCmd = &cobra.Command{
	Use:   "privbind",
	Short: "Allow binding to privileged ports",
	Long: `Configure the local system to allow importing remote services on
privileged ports (<1024).`,
	Run: func(cmd *cobra.Command, args []string) {
		err := runner.Run(&runner.AgentPrivbind{Base: runner.Base{ConfigFile: configFile}}, args)
		if err != nil {
			log.Fatalf("%v", err)
		}
	},
}

func init() {
	agentCmd.AddCommand(agentPrivbindCmd)
}
