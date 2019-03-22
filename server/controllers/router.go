package controllers

import (
	"github.com/getupandgo/gocache/common/cache"
	"github.com/getupandgo/gocache/server/controllers/pagecache"
	"github.com/getupandgo/gocache/server/middlewares"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func InitRouter(cc cache.CacheClient) *gin.Engine {
	cacheCtrl := pagecache.Init(cc)

	maxPageSize := viper.GetInt64("limits.max_record_size")

	r := gin.New()

	r.Use(middlewares.AppErrorReporter())
	r.Use(middlewares.RequestSizeLimiter(maxPageSize))

	cacheRouter := r.Group("/cache")
	cacheRouter.GET("", cacheCtrl.GetPage)
	cacheRouter.PUT("", cacheCtrl.UpsertPage)
	cacheRouter.DELETE("", cacheCtrl.DeletePage)

	cacheRouter.GET("/top", cacheCtrl.GetTopPages)

	return r

}
