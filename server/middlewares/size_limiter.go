package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func BodySizeLimiter(maxRecordSize int64) gin.HandlerFunc {
	return func(c *gin.Context) {

		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxRecordSize)

		c.Next()
	}
}
