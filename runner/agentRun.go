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

package runner

import (
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"

	"github.com/cmars/ormesh/agent"
	"github.com/cmars/ormesh/config"
)

type AgentRun struct {
	Base
}

func (r *AgentRun) Run(args []string) error {
	if exportsValue := os.Getenv("ORMESH_EXPORTS"); exportsValue != "" {
		exportAdd := &ExportAdd{Base: r.Base}
		exports := strings.Split(exportsValue, ";")
		for i := range exports {
			exportArgs := strings.Split(exports[i], " ")
			err := exportAdd.Run(exportArgs)
			if err != nil {
				return errors.Wrapf(err, "export add %q", exports[i])
			}
		}
	}
	if clientsValue := os.Getenv("ORMESH_CLIENTS"); clientsValue != "" {
		clientAdd := &ClientAdd{Base: r.Base}
		clients := strings.Split(clientsValue, ";")
		for i := range clients {
			err := clientAdd.Run([]string{clients[i]})
			if err != nil {
				return errors.Wrapf(err, "client add %q", clients[i])
			}
		}
	}
	err := r.WithConfig(func(cfg *config.Config) error {
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
		err = refresh(cfg)
		if err != nil {
			return errors.WithStack(err)
		}

		if cfg.Node.Agent.UseTorBrowser {
			var nImports int
			for _, remote := range cfg.Node.Remotes {
				nImports += len(remote.Imports)
			}
			if nImports == 0 {
				log.Println("no imports to forward, exiting")
				return nil
			}
		}

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
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}
