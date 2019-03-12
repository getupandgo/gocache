package main

import (
	"fmt"
	"github.com/getupandgo/gocache/server/controllers"
	"github.com/getupandgo/gocache/utils/cache"
	"github.com/getupandgo/gocache/utils/config"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func main() {
	conf, err := config.Get()
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("Failed to get config file")
	}

	rd := cache.Init(conf)

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

	log.Info().Msgf("starting cache server on port %d", httpPort)
}
