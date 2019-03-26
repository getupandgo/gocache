package utils

import (
	"time"

	"github.com/getupandgo/gocache/common/cache"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func WatchExpiredRecords(cc cache.Page) error {
	pollInterv := viper.GetString("cache.ttl_eviction_interval")

	pollDuration, err := time.ParseDuration(pollInterv)
	if err != nil {
		return err
	}

	ticker := time.NewTicker(pollDuration)

	go func() {
		for t := range ticker.C {
			_ = t

			if _, err := cc.Expire(); err != nil {
				log.Info().
					Err(err).
					Msg("Failed to delete expired records")
			}
		}
	}()

	return nil
}
