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
	"os/signal"
	"syscall"

	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/cmars/ormesh/agent"
	"github.com/cmars/ormesh/config"
)

// agentRunCmd represents the agentRun command
var agentRunCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the ormesh agent",
	Long: `The agent launches and operate a tor subprocess, implementing the configured
service policies. Configuration is automatically refreshed and applied when the
ormesh configuration file is modified or a SIGHUP received. This command will
not exit until an interrupt signal is received or an error is encountered.`,
	Run: func(cmd *cobra.Command, args []string) {
		withConfig(func(cfg *config.Config) error {
			a, err := agent.New(cfg)
			if err != nil {
				return errors.Wrap(err, "failed to initialize agent")
			}
			err = a.Start()
			if err != nil {
				return errors.Wrap(err, "failed to start agent")
			}
			defer a.Stop()

			refresh := func(cfg *config.Config) error {
				err = a.UpdateServices(&cfg.Node.Service)
				if err != nil {
					return errors.Wrap(err, "failed to configure hidden services")
				}
				err = a.UpdateRemotes(&cfg.Node)
				if err != nil {
					return errors.Wrap(err, "failed to configure remotes")
				}
				return nil
			}
			refresh(cfg)

			watcher, err := fsnotify.NewWatcher()
			if err != nil {
				return errors.WithStack(err)
			}
			defer watcher.Close()
			err = watcher.Add(cfg.Path)
			if err != nil {
				return errors.WithStack(err)
			}

			exitSignal := make(chan os.Signal, 1)
			signal.Notify(exitSignal, syscall.SIGINT, syscall.SIGTERM)

			refreshSignal := make(chan os.Signal, 1)
			signal.Notify(exitSignal, syscall.SIGHUP)

			for {
				select {
				case s := <-exitSignal:
					return errors.Errorf("exit on signal %v", s)
				case <-refreshSignal:
					err = refresh(cfg)
					if err != nil {
						return errors.WithStack(err)
					}
				case ev := <-watcher.Events:
					if ev.Op&fsnotify.Write == fsnotify.Write {
						cfg, err = config.ReadFile(cfg.Path)
						if err != nil {
							return errors.WithStack(err)
						}
						log.Printf("configuration changed")
						err = refresh(cfg)
						if err != nil {
							return errors.WithStack(err)
						}
					}
				}
			}
		})
	},
}

func init() {
	agentCmd.AddCommand(agentRunCmd)
}
