package utils

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

type DBOptions struct {
	Connection         string
	TopRecordsCount    int64
	ExpirationInterval string
	MaxCount           int64
	MaxSize            int64
}

func ReadDbOptions() DBOptions {
	connString := viper.GetString("redis.host") + ":" + viper.GetString("redis.port")

	return DBOptions{
		connString,
		viper.GetInt64("cache.top_records_count"),
		viper.GetString("cache.expiration_interval"),
		viper.GetInt64("limits.max_count"),
		viper.GetInt64("limits.max_size"),
	}
}

func ReadConfig(path string) {
	viper.AutomaticEnv()
	viper.AddConfigPath(path)
	viper.SetConfigName("default")
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatal().
			Err(err).
			Msg("Failed to get config file")
	}
}
