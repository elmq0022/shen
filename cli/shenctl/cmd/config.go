package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

func InitViper() {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(
			os.Stderr,
			"Warning: could not determine home directory, config file support disabled: %v\n",
			err,
		)
		return
	}

	configPath := filepath.Join(home, ".shenctl")

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
