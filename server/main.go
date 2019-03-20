package main

import (
	"fmt"

	"github.com/getupandgo/gocache/common/cache"
	"github.com/getupandgo/gocache/common/utils"
	"github.com/getupandgo/gocache/server/controllers"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func main() {
	utils.ReadConfig()

	rd, err := cache.Init()
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("Failed to init cache client")
	}

	if !viper.GetBool("http_debug") {
		gin.SetMode(gin.ReleaseMode)
	}

	r := controllers.InitRouter(rd)

	httpPort := viper.GetInt("server.port")

	if err = r.Run(fmt.Sprintf(":%d", httpPort)); err != nil {
		log.Fatal().
			Err(err).
			Msg("Failed to start server")
	}

	if err := utils.WatchExpiredRecords(rd); err != nil {
		log.Fatal().
			Err(err).
			Msg("Failed to start expiration watcher")
	}

	log.Info().Msgf("starting cache server on port %d", httpPort)
}
