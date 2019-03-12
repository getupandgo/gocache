package pagecache

import (
	"github.com/getupandgo/gocache/utils/cache"
	"github.com/gin-gonic/gin"
	"net/http"
)

type (
	CacheController struct {
		cacheClient cache.CacheClient
	}

	SavePageRequest struct {
		Url string
	}
)

func Init(cc cache.CacheClient) *CacheController {
	return &CacheController{cc}
}

func (ctrl *CacheController) GetPage(c *gin.Context) {

}

func (ctrl *CacheController) UpsertPage(c *gin.Context) {
}

func (ctrl *CacheController) DeletePage(c *gin.Context) {
}

func (ctrl *CacheController) GetTopPages(c *gin.Context) {
}
