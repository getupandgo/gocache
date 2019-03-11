package pagecache

import (
	"github.com/getupandgo/gocache/utils/cache"
	"github.com/gin-gonic/gin"
)

type CacheController struct {
	cacheClient *cache.RedisClient
}

func Init(cc *cache.RedisClient) CacheController {
	return CacheController{cc}
}

func (ctrl CacheController) GetPage(c *gin.Context) {
}

func (ctrl CacheController) UpsertPage(c *gin.Context) {
}
