package pagecache

import (
	"io/ioutil"
	"mime/multipart"
	"net/http"

	"github.com/getupandgo/gocache/common/cache"
	"github.com/getupandgo/gocache/common/structs"
	"github.com/gin-gonic/gin"
)

type CacheController struct {
	db cache.Page
}

func Init(cc cache.Page) *CacheController {
	return &CacheController{cc}
}

func (ctrl *CacheController) GetPage(c *gin.Context) {
	pg := c.Query("url")

	cont, err := ctrl.db.Get(pg)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, cont)
}

func (ctrl *CacheController) UpsertPage(c *gin.Context) {
	pageURL := c.PostForm("url")
	fh, err := c.FormFile("content")
	if err != nil {
		c.Error(err)
		return
	}

	cont, err := ReadMultipart(fh)
	if err != nil {
		c.Error(err)
		return
	}

	totalDataSize := len(cont) + len(pageURL)

	if err := ctrl.db.Upsert(&structs.Page{pageURL, cont, totalDataSize}); err != nil {
		c.Error(err)
		return
	}

	c.String(http.StatusOK, pageURL)
}

func (ctrl *CacheController) DeletePage(c *gin.Context) {
	removePage := &structs.RemovePageBody{}

	if err := c.BindJSON(removePage); err != nil {
		c.Error(err)
		return
	}

	_, err := ctrl.db.Remove(removePage.URL)
	if err != nil {
		c.Error(err)
		return
	}

	c.Status(http.StatusOK)

}

func (ctrl *CacheController) GetTopPages(c *gin.Context) {
	top, err := ctrl.db.Top()
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, top)
}

func ReadMultipart(cont *multipart.FileHeader) ([]byte, error) {
	src, err := cont.Open()
	if err != nil {
		return nil, err
	}

	defer src.Close()

	b, err := ioutil.ReadAll(src)
	if err != nil {
		return nil, err
	}

	return b, nil
}
