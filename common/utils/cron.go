package utils

import (
	"time"

	"github.com/getupandgo/gocache/common/cache"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func WatchExpiredRecords(cc cache.Page) error {
	ticker := time.NewTicker(1 * time.Second)

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

	pollInterv := viper.GetInt64("limits.expire_interval")
	time.Sleep(10 * time.Duration(pollInterv))
	ticker.Stop()

	return nil
}
