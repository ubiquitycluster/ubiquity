/*
Copyright © 2026 Ubiquity Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/ubiquitycluster/ubiquity/blob/main/LICENSE

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "ubiquity",
	Short: "HPC Cluster Lifecycle Platform",
	Long: `Ubiquity provisions, operates, and updates self-hosted HPC clusters
using Infrastructure as Code and GitOps principles.

Complete documentation: https://ubiquitycluster.github.io/ubiquity`,
}

// Execute adds all child commands to the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().String("config", "", "config file (default is $HOME/.ubiquity/.ubiquity.yaml)")
	rootCmd.PersistentFlags().String("env", "sandbox", "deployment environment (sandbox, dev, prod)")

	// Bind Viper to flags so UBQUITY_ENV env var and --env flag both work
	viper.BindPFlag("env", rootCmd.PersistentFlags().Lookup("env"))
	viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config"))

	// Environment variable mapping: UBQUITY_ENV -> env
	viper.SetEnvPrefix("UBQUITY")
	viper.AutomaticEnv()

	// Version flag (registered here, not in initConfig, so tests can see it)
	rootCmd.Flags().BoolP("version", "v", false, "print version information")
	rootCmd.Run = func(cmd *cobra.Command, args []string) {
		if v, _ := cmd.Flags().GetBool("version"); v {
			fmt.Printf("Ubiquity CLI %s (commit: %s, built: %s)\n", Version, Commit, Date)
			return
		}
		cmd.Help()
	}
}

// initConfig reads the config file and sets up Viper.
func initConfig() {
	cfgFile := viper.GetString("config")
	if cfgFile != "" {
		// Use explicit config file
		viper.SetConfigFile(cfgFile)
	} else {
		// Search default paths
		home, err := os.UserHomeDir()
		if err == nil {
			viper.AddConfigPath(filepath.Join(home, ".ubiquity"))
		}
		viper.AddConfigPath(".")
		viper.SetConfigName(".ubiquity")
		viper.SetConfigType("yaml")
	}

	// Read config — ignore error if file doesn't exist
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			fmt.Fprintf(os.Stderr, "Warning: error reading config: %v\n", err)
		}
	}
}

// Env returns the current environment from viper (--env flag or UBQUITY_ENV).
func Env() string {
	return viper.GetString("env")
}