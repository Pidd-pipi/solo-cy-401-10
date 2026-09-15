package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"

	"github.com/gigmatch/gigmatch/internal/constants"
	"github.com/gigmatch/gigmatch/internal/util"
)

// ErrorHandler recovers panics and returns a unified JSON error.
func ErrorHandler(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("panic recovered",
					"panic", r,
					"path", c.Request.URL.Path,
					"stack", string(debug.Stack()))
				c.AbortWithStatusJSON(http.StatusInternalServerError, util.Response{
					Code:    constants.CodeInternal,
					Message: constants.MsgInternalError,
				})
			}
		}()
		c.Next()
	}
}
