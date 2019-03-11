package config

import (
	"github.com/spf13/viper"
)

const defaultConfigPath = "config"

func GetConfig() (*viper.Viper, error) {
	v := viper.New()

	v.AutomaticEnv()
	v.AddConfigPath(defaultConfigPath)
	v.SetConfigName("default")
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}
	return v, nil
}
