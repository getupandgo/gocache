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
	token := &SavePageRequest{}

	if err := c.BindJSON(token); err != nil {
		c.Error(err)
		return
	}

	err, content := getHTML(token.Url)
	if err != nil {
		c.Error(err)
		return
	}

	if err := ctrl.cacheClient.UpsertPage(&cache.PageMsg{token.Url, string(content)}); err != nil {
		c.Error(err)
		return
	}

	c.String(http.StatusOK, token.Url)
}

func (ctrl *CacheController) DeletePage(c *gin.Context) {
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
