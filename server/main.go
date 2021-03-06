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

const defaultConfigPath = "config"

func main() {
	utils.ReadConfig(defaultConfigPath)

	rd, err := cache.Init(utils.ReadDbOptions())
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("Failed to init cache client")
	}

	if !viper.GetBool("http_debug") {
		gin.SetMode(gin.ReleaseMode)
	}

	maxReqSize := viper.GetInt64("limits.record.max_size")
	defaultTTL := viper.GetString("limits.record.ttl")

	r := controllers.InitRouter(rd, maxReqSize, defaultTTL)

	httpPort := viper.GetInt("server.port")

	log.Info().Msgf("starting cache server on port %d", httpPort)

	if err = r.Run(fmt.Sprintf(":%d", httpPort)); err != nil {
		log.Fatal().
			Err(err).
			Msg("Failed to start server")
	}
}
