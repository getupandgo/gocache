package middlewares

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func RequestSizeLimiter(maxSize int64) gin.HandlerFunc {
	return func(c *gin.Context) {

		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxSize)

		c.Next()
	}
}
