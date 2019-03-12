package controllers

import (
	"github.com/getupandgo/gocache/server/controllers/pagecache"
	"github.com/getupandgo/gocache/server/middlewares"
	"github.com/getupandgo/gocache/utils/cache"
	"github.com/gin-gonic/gin"
)

func InitRouter(cc cache.CacheClient) *gin.Engine {
	cacheCtrl := pagecache.Init(cc)

	r := gin.New()

	r.Use(middlewares.AppErrorReporter())

	cacheRouter := r.Group("/cache")
	cacheRouter.GET("", cacheCtrl.GetPage)
	cacheRouter.PUT("", cacheCtrl.UpsertPage)
	cacheRouter.DELETE("", cacheCtrl.DeletePage)

	topRouter := r.Group("/top")
	topRouter.GET("", cacheCtrl.GetTopPages)

	return r
}
