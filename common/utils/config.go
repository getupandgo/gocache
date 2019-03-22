package utils

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

const defaultConfigPath = "config"

func ReadConfig() {
	viper.AutomaticEnv()
	viper.AddConfigPath(defaultConfigPath)
	viper.SetConfigName("default")
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatal().
			Err(err).
			Msg("Failed to get config file")
	}
}
