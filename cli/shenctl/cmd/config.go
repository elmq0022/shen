/*
Copyright © 2026 Aaron Elmquist
*/

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func getConfigPath() string {
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		configHome = filepath.Join(home, ".config")
	}
	return filepath.Join(configHome, "shenctl")
}

func InitViper() {
	configPath := getConfigPath()
	if configPath == "" {
		fmt.Fprintf(
			os.Stderr,
			"Warning: could not determine config directory, config file support disabled\n",
		)
		return
	}

	viper.AddConfigPath(configPath)
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			fmt.Fprintf(os.Stderr, "Warning: error reading config file: %v\n", err)
		}
		// Config file not found is OK - continue with defaults
	}
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage shenctl configuration settings",
	Long:  `Manage shenctl configuration settings stored in $XDG_CONFIG_HOME/shenctl/config.toml (default: ~/.config/shenctl/config.toml).`,
}

var listConfigCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configuration values",
	Long:  `Display all configuration values currently set in the shenctl configuration file.`,
	Run: func(cmd *cobra.Command, args []string) {
		settings := viper.AllSettings()
		if len(settings) == 0 {
			fmt.Println("No configuration values set")
			return
		}

		keys := make([]string, 0, len(settings))
		for k := range settings {
			keys = append(keys, k)
		}
		slices.Sort(keys)

		for _, key := range keys {
			fmt.Printf("%s=%v\n", key, settings[key])
		}
	},
}

var getConfigCmd = &cobra.Command{
	Use:   "get",
	Short: "Retrieve a configuration value",
	Long:  `Retrieve and display the value of a specific configuration setting from the shenctl configuration file.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Fprintln(os.Stderr, "Error: key required")
			os.Exit(1)
		}

		key := args[0]
		value := viper.Get(key)

		if !viper.IsSet(key) || value == nil {
			fmt.Fprintf(os.Stderr, "Error: key '%s' not found in config\n", key)
			os.Exit(1)
		}

		fmt.Println(value)
	},
}

var setConfigCmd = &cobra.Command{
	Use:   "set",
	Short: "Set a configuration value",
	Long:  `Set or update a configuration value in the shenctl configuration file.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 2 {
			fmt.Fprintln(os.Stderr, "Set requires exactly one key and one value.")
			os.Exit(1)
		}
		viper.Set(args[0], args[1])
		if err := writeConfig(); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing config: %v\n", err)
			os.Exit(1)
		}
	},
}

var unsetConfigCmd = &cobra.Command{
	Use:   "unset",
	Short: "Remove a configuration value",
	Long:  `Remove a configuration value from the shenctl configuration file, reverting it to its default.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Fprintln(os.Stderr, "Error: key required")
			os.Exit(1)
		}

		key := args[0]

		settings := viper.AllSettings()

		if _, exists := settings[key]; !exists {
			fmt.Fprintf(os.Stderr, "Key '%s' not found in config\n", key)
			os.Exit(1)
		}

		delete(settings, key)

		configPath := getConfigPath()
		if configPath == "" {
			fmt.Fprintln(os.Stderr, "Error: could not determine config directory")
			os.Exit(1)
		}
		configFile := filepath.Join(configPath, "config.toml")

		newViper := viper.New()
		newViper.SetConfigFile(configFile)
		newViper.SetConfigType("toml")

		for k, v := range settings {
			newViper.Set(k, v)
		}

		if err := newViper.WriteConfig(); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing config: %v\n", err)
			os.Exit(1)
		}
	},
}

func writeConfig() error {
	if err := viper.WriteConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			configPath := getConfigPath()
			if configPath == "" {
				return fmt.Errorf("could not determine config directory")
			}
			if err := os.MkdirAll(configPath, 0755); err != nil {
				return err
			}
			return viper.SafeWriteConfig()
		}
		return err
	}
	return nil
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(
		listConfigCmd,
		getConfigCmd,
		setConfigCmd,
		unsetConfigCmd,
	)
}
