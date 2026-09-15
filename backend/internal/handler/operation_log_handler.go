package handler

import (
	"log/slog"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/gigmatch/gigmatch/internal/service"
	"github.com/gigmatch/gigmatch/internal/util"
)

// OperationLogHandler exposes operation log endpoints.
type OperationLogHandler struct {
	svc    *service.OperationLogService
	logger *slog.Logger
}

// NewOperationLogHandler builds an OperationLogHandler.
func NewOperationLogHandler(svc *service.OperationLogService, logger *slog.Logger) *OperationLogHandler {
	return &OperationLogHandler{svc: svc, logger: logger}
}

// List handles GET /operation-logs?limit=.
func (h *OperationLogHandler) List(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	logs, err := h.svc.List(limit)
	if err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, logs)
}
