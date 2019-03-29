package controllers

import (
	"github.com/getupandgo/gocache/common/cache"
	"github.com/getupandgo/gocache/server/controllers/pagecache"
	"github.com/getupandgo/gocache/server/middlewares"
	"github.com/gin-gonic/gin"
)

func InitRouter(cc cache.Page, maxRequestSize int64, defaultTTL string) *gin.Engine {
	cacheCtrl := pagecache.Init(cc, defaultTTL)

	r := gin.New()

	r.Use(middlewares.AppErrorReporter())
	r.Use(middlewares.BodySizeLimiter(maxRequestSize))

	cacheRouter := r.Group("/cache")
	cacheRouter.POST("/get", cacheCtrl.GetPage)
	cacheRouter.PUT("/upsert", cacheCtrl.UpsertPage)
	cacheRouter.POST("/delete", cacheCtrl.DeletePage)
	cacheRouter.GET("/top", cacheCtrl.GetTopPages)

	return r

}
