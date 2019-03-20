package utils

import (
	"github.com/getupandgo/gocache/common/cache"
	"github.com/robfig/cron"
	"github.com/rs/zerolog/log"
)

func WatchExpiredRecords(cc cache.CacheClient) error {
	c := cron.New()
	err := c.AddFunc("@every 1m", func() {
		if _, err := cc.RemoveExpiredRecords(); err != nil {
			log.Info().
				Err(err).
				Msg("Failed to delete expired records")
		}
	})

	if err != nil {
		return err
	}

	c.Start()

	return nil
}
