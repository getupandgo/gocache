package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type Error struct {
	Status  int
	Message string
}

func AppErrorReporter() gin.HandlerFunc {
	return func(c *gin.Context) {
		// try
		c.Next()
		// catch
		if err := c.Errors.Last(); err != nil {
			var formattedError Error

			switch err.Type {
			case gin.ErrorTypePublic:
				log.Info().
					Err(err).
					Msg("")

				formattedError = Error{http.StatusBadRequest, err.Error()}

			case gin.ErrorTypeBind:
				log.Info().
					Err(err).
					Msg("")

				formattedError = Error{http.StatusBadRequest, "Invalid param"}

			case gin.ErrorTypePrivate:
				log.Error().
					Err(err).
					Msg("")

				formattedError = Error{http.StatusInternalServerError, "Server error"}

			default:
				log.Error().
					Err(err).
					Msg("Unhandled error")

				formattedError = Error{http.StatusInternalServerError, "Server error"}
			}

			c.AbortWithStatusJSON(formattedError.Status, formattedError.Message)
		}
	}
}
