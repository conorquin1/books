package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Init initializes viper for environment variable management.
func Init() error {
	viper.SetConfigName("config")
	viper.AddConfigPath("/opt/books/conf/")
	viper.AddConfigPath(".")
	viper.SetEnvPrefix("books")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		// Config file not found is OK - we'll use env vars
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("viper.ReadInConfig: %v", err)
		}
	}

	return nil
}
