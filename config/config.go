package config

import (
	"github.com/spf13/viper"
)

// LoadConfig loads the config file from the given path
func LoadBetrayalConfig() {
	viper.SetConfigFile(".env")
	viper.ReadInConfig()
}
