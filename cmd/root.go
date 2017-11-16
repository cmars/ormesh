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
	"path/filepath"

	"github.com/cmars/ormesh/config"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var cfgFile string

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "ormesh",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.ormesh.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	RootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile == "" {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			log.Fatalf("failed to locate home directory: %v", err)
		}
		ormeshDir := filepath.Join(home, ".ormesh")
		if err := os.MkdirAll(ormeshDir, 0700); err != nil {
			log.Fatalf("failed to create %q: %v", ormeshDir, err)
		}
		cfgFile = filepath.Join(ormeshDir, "config")
	}
}

func withConfig(f func(*config.Config) error) {
	cfg, err := config.ReadFile(cfgFile)
	if os.IsNotExist(errors.Cause(err)) {
		cfg = config.NewFile(cfgFile)
	} else if err != nil {
		log.Fatalf("%v", err)
	}
	err = f(cfg)
	if err != nil {
		log.Fatalf("%v", err)
	}
}

func withConfigForUpdate(f func(*config.Config) error) {
	cfg, err := config.ReadFile(cfgFile)
	if os.IsNotExist(errors.Cause(err)) {
		cfg = config.NewFile(cfgFile)
	} else if err != nil {
		log.Fatalf("%v", err)
	}
	err = f(cfg)
	if err != nil {
		log.Fatalf("%v", err)
	}
	err = config.WriteFile(cfg, cfgFile)
	if err != nil {
		log.Fatalf("%v", err)
	}
}
