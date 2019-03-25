package utils

import (
	"time"

	"github.com/getupandgo/gocache/common/cache"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func WatchExpiredRecords(cc cache.Page) error {
	pollInterv := viper.GetInt64("cache.ttl_eviction_interval")
	ticker := time.NewTicker(time.Duration(pollInterv))

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

	time.Sleep(10 * time.Duration(pollInterv))

	return nil
}
