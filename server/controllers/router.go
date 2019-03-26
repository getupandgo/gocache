package controllers

import (
	"github.com/getupandgo/gocache/common/cache"
	"github.com/getupandgo/gocache/server/controllers/pagecache"
	"github.com/getupandgo/gocache/server/middlewares"
	"github.com/gin-gonic/gin"
)

func InitRouter(cc cache.Page, maxRequestSize int64) *gin.Engine {
	cacheCtrl := pagecache.Init(cc)

	r := gin.New()

	r.Use(middlewares.AppErrorReporter())
	r.Use(middlewares.BodySizeLimiter(maxRequestSize))

	cacheRouter := r.Group("/cache")
	cacheRouter.POST("", cacheCtrl.GetPage)
	cacheRouter.PUT("", cacheCtrl.UpsertPage)
	cacheRouter.DELETE("", cacheCtrl.DeletePage)

	cacheRouter.GET("/top", cacheCtrl.GetTopPages)

	return r

}
