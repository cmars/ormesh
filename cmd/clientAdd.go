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
	"encoding/json"
	"fmt"
	"os"

	"github.com/mdp/qrterminal"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/cmars/ormesh/agent"
	"github.com/cmars/ormesh/config"
)

var displayQR bool

// clientAddCmd represents the clientAdd command
var clientAddCmd = &cobra.Command{
	Use:   "add <client name>",
	Short: "Add a client authorization",
	Long: `Create an auth token allowing a client to access exported services. The token
should be securely transmitted to the client.`,
	Example: `
  $ ormesh client add my-MacBook
  Y5Cfw7A5RhP8Rd7xGYfD8N4oyEBpBWNR+6Qkgrbepk0=

  Then paste this token as an argument to 'ormesh remote add':

  $ ormesh remote add my-server Y5Cfw7A5RhP8Rd7xGYfD8N4oyEBpBWNR+6Qkgrbepk0=`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		withConfigForUpdate(func(cfg *config.Config) error {
			client, err := addClient(cfg, args[0])
			if err != nil {
				return errors.WithStack(err)
			}
			if displayQR {
				qrDoc := struct {
					AuthCookieValue string `json:"auth_cookie_value"`
					Domain          string `json:"domain"`
				}{
					AuthCookieValue: client.Auth,
					Domain:          client.Address,
				}
				qrText, err := json.Marshal(&qrDoc)
				if err != nil {
					return errors.WithStack(err)
				}
				qrterminal.Generate(string(qrText), qrterminal.H, os.Stdout)
			} else {
				fmt.Printf("%s %s\n", client.Address, client.Auth)
			}
			return nil
		})
	},
}

func addClient(cfg *config.Config, clientName string) (*config.Client, error) {
	if !IsValidClientName(clientName) {
		return nil, errors.Errorf("invalid client name %q", clientName)
	}
	index := -1
	for i := range cfg.Node.Service.Clients {
		if cfg.Node.Service.Clients[i].Name == clientName {
			index = i
			break
		}
	}
	if index < 0 {
		index = len(cfg.Node.Service.Clients)
		cfg.Node.Service.Clients = append(cfg.Node.Service.Clients, config.Client{
			Name: clientName,
		})
	}
	a, err := agent.New(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize agent")
	}
	err = a.Start()
	if err != nil {
		return nil, errors.Wrap(err, "failed to start agent")
	}
	defer a.Stop()
	err = a.UpdateServices(&cfg.Node.Service)
	if err != nil {
		return nil, errors.Wrap(err, "failed to update tor hidden services")
	}
	address, clientAuth, err := a.ClientAccess(clientName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read tor client auth")
	}
	cfg.Node.Service.Clients[index].Address = address
	cfg.Node.Service.Clients[index].Auth = clientAuth
	return &cfg.Node.Service.Clients[index], nil
}

func init() {
	clientAddCmd.Flags().BoolVarP(&displayQR, "qr", "", false, "Display Orbot client cookie QR code")
	clientCmd.AddCommand(clientAddCmd)
}
