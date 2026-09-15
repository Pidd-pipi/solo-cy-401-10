package middleware

import (
	"log/slog"

	"github.com/gin-gonic/gin"

	"github.com/gigmatch/gigmatch/internal/service"
)

// AuditLogger records non-GET mutations at the HTTP layer; business-level
// entries are additionally written by services.
func AuditLogger(logs *service.OperationLogService, logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if c.Request.Method == "GET" || c.Request.Method == "OPTIONS" || c.Writer.Status() >= 400 {
			return
		}
		user := GetCurrentUser(c)
		logs.Record(user.ID, user.Name, c.Request.Method+" "+c.Request.URL.Path, "http", 0, c.Request.URL.Path)
		_ = logger
	}
}
