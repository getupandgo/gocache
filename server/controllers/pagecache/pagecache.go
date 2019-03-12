package pagecache

import (
	"github.com/getupandgo/gocache/utils/cache"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
)

type (
	CacheController struct {
		cacheClient cache.CacheClient
	}

	RemovePageRequest struct {
		Url string
	}
)

func Init(cc cache.CacheClient) *CacheController {
	return &CacheController{cc}
}

func (ctrl *CacheController) GetPage(c *gin.Context) {

}

func (ctrl *CacheController) UpsertPage(c *gin.Context) {
	newPage := &RemovePageRequest{}

	if err := c.BindJSON(newPage); err != nil {
		c.Error(err)
		return
	}

	err, content := getHTML(newPage.Url)
	if err != nil {
		c.Error(err)
		return
	}

	if err := ctrl.cacheClient.UpsertPage(&cache.PageMsg{newPage.Url, string(content)}); err != nil {
		c.Error(err)
		return
	}

	c.String(http.StatusOK, newPage.Url)
}

func (ctrl *CacheController) DeletePage(c *gin.Context) {
	removePage := &RemovePageRequest{}

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

func getHTML(url string) (error, []byte) {
	resp, err := http.Get(url)
	if err != nil {
		return err, nil
	}

	defer resp.Body.Close()

	html, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err, nil
	}

	return nil, html
}
