package main

import (
	"fmt"
	"github.com/getupandgo/gocache/common/cache"
	"github.com/getupandgo/gocache/common/config"
	"github.com/getupandgo/gocache/server/controllers"
	"github.com/gin-gonic/gin"
	"github.com/robfig/cron"
	"github.com/rs/zerolog/log"
)

func main() {
	conf, err := config.Get()
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("Failed to get config file")
	}

	rd, err := cache.Init(conf)
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("Failed to init cache client")
	}

	if !conf.GetBool("http_debug") {
		gin.SetMode(gin.ReleaseMode)
	}

	r := controllers.InitRouter(rd)

	httpPort := conf.GetInt("server.port")

	if err = r.Run(fmt.Sprintf(":%d", httpPort)); err != nil {
		log.Fatal().
			Err(err).
			Msg("Failed to start server")
	}

	if err := watchExpiredRecords(rd); err != nil {
		log.Fatal().
			Err(err).
			Msg("Failed to start expiration watcher")
	}

	log.Info().Msgf("starting cache server on port %d", httpPort)
}

func watchExpiredRecords(cc cache.CacheClient) error {
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
