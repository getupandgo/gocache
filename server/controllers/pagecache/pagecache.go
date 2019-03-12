package pagecache

import (
	"github.com/getupandgo/gocache/utils/cache"
	"github.com/getupandgo/gocache/utils/structs"
	"github.com/gin-gonic/gin"
	"net/http"
)

type CacheController struct {
	cacheClient cache.CacheClient
}

func Init(cc cache.CacheClient) *CacheController {
	return &CacheController{cc}
}

func (ctrl *CacheController) GetPage(c *gin.Context) {

}

func (ctrl *CacheController) UpsertPage(c *gin.Context) {
	newPage := &structs.Page{}

	if err := c.BindJSON(newPage); err != nil {
		c.Error(err)
		return
	}

	if err := ctrl.cacheClient.UpsertPage(newPage); err != nil {
		c.Error(err)
		return
	}

	c.String(http.StatusOK, newPage.Url)
}

func (ctrl *CacheController) DeletePage(c *gin.Context) {
	removePage := &structs.RemovePageBody{}

	if err := c.BindJSON(removePage); err != nil {
		c.Error(err)
		return
	}

	ctrl.cacheClient.RemovePage(removePage.Url)
}

func (ctrl *CacheController) GetTopPages(c *gin.Context) {
	err, top := ctrl.cacheClient.GetTopPages()
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, top)
}
