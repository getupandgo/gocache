package pagecache

import (
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"strconv"

	"github.com/spf13/viper"

	"github.com/go-redis/redis"

	"github.com/getupandgo/gocache/common/utils"

	"github.com/getupandgo/gocache/common/cache"
	"github.com/getupandgo/gocache/common/structs"
	"github.com/gin-gonic/gin"
)

type CacheController struct {
	db         cache.Page
	defaultTTL string
}

func Init(cc cache.Page) *CacheController {
	defaultTTL := viper.GetString("limits.record.ttl")

	return &CacheController{cc, defaultTTL}
}

func (ctrl *CacheController) GetPage(c *gin.Context) {
	pg, present := c.GetPostForm("url")
	if !present {
		c.JSON(http.StatusBadRequest, "No URL provided")
		return
	}

	cont, err := ctrl.db.Get(pg)
	if err == redis.Nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	} else if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, string(cont))
}

func (ctrl *CacheController) UpsertPage(c *gin.Context) {
	pageURL, present := c.GetPostForm("url")
	if !present {
		c.JSON(http.StatusBadRequest, "No URL provided")
		return
	}

	reqTTL, present := c.GetPostForm("ttl")
	var pageTTL int64
	var err error

	if !present || reqTTL == "" {
		pageTTL, err = utils.CalculateTTLFromNow(ctrl.defaultTTL)
	} else {
		pageTTL, err = strconv.ParseInt(reqTTL, 10, 64)
	}

	fh, err := c.FormFile("content")
	if err != nil {
		c.JSON(http.StatusBadRequest, "No content provided")
		return
	}

	content, err := readMultipart(fh)
	if err != nil {
		c.JSON(http.StatusBadRequest, "Cannot read page content")
		return
	}

	totalDataSize := len(content) + len(pageURL)

	_, err = ctrl.db.Upsert(
		&structs.Page{
			pageURL, content, pageTTL, totalDataSize,
		})
	if err != nil {
		c.Error(err)
		return
	}

	c.String(http.StatusOK, pageURL)
}

func (ctrl *CacheController) DeletePage(c *gin.Context) {
	pageURL, present := c.GetPostForm("url")
	if !present {
		c.JSON(http.StatusBadRequest, "No URL provided")
		return
	}

	bytesRemoved, err := ctrl.db.Remove(pageURL)
	if err == redis.Nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	} else if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, bytesRemoved > 0)

}

func (ctrl *CacheController) GetTopPages(c *gin.Context) {
	top, err := ctrl.db.Top()
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, top)
}

func readMultipart(cont *multipart.FileHeader) ([]byte, error) {
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
