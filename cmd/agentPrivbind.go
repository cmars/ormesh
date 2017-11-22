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
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

// agentPrivbindCmd represents the agentPrivbind command
var agentPrivbindCmd = &cobra.Command{
	Use:   "privbind",
	Short: "Allow binding to privileged ports",
	Long: `Configure the local system to allow importing remote services on
privileged ports (<1024).`,
	Run: func(cmd *cobra.Command, args []string) {
		binaryPath, err := filepath.Abs(os.Args[0])
		if err != nil {
			log.Fatalf("%v", err)
		}
		if os.Getuid() == 0 {
			cmd := exec.Command("setcap", "cap_net_bind_service=+ep", binaryPath)
			log.Printf("%#v", cmd)
			err := cmd.Run()
			if err != nil {
				log.Fatalf("setcap failed: %v", err)
			}
		} else {
			cmd := exec.Command("/bin/sh", "-c",
				fmt.Sprintf("sudo setcap 'cap_net_bind_service=+ep' %s", binaryPath))
			log.Printf("%#v", cmd)
			err := cmd.Run()
			if err != nil {
				log.Fatalf("setcap failed: %v", err)
			}
		}
	},
}

func init() {
	agentCmd.AddCommand(agentPrivbindCmd)
}
