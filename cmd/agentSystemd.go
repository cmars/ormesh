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
	"os"
	"path/filepath"
	"text/template"

	"github.com/spf13/cobra"
)

var serviceTemplate = template.Must(template.New("systemd").Parse(`
[Unit]
Description=Onion-routed mesh

[Service]
ExecStart={{.BinaryPath}} agent run
Restart=always
{{ if .Username -}}
User={{ .Username }}
{{- else -}}
User=ubuntu
{{- end }}

[Install]
WantedBy=multi-user.target
`))

var serviceUser string

// agentSystemdCmd represents the agentSystemd command
var agentSystemdCmd = &cobra.Command{
	Use:   "systemd",
	Short: "Generate a systemd service that operates ormesh",
	Run: func(cmd *cobra.Command, args []string) {
		binaryPath, err := filepath.Abs(os.Args[0])
		if err != nil {
			log.Fatalf("%v", err)
		}
		err = serviceTemplate.Execute(os.Stdout, struct {
			BinaryPath string
			Username   string
		}{
			BinaryPath: binaryPath,
			Username:   serviceUser,
		})
		if err != nil {
			log.Fatalf("%v", err)
		}
	},
}

func init() {
	agentSystemdCmd.Flags().StringVarP(&serviceUser, "user", "u", "", "User to run ormesh as")
	agentCmd.AddCommand(agentSystemdCmd)
}
