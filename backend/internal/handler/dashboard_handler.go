package handler

import (
	"log/slog"

	"github.com/gin-gonic/gin"

	"github.com/gigmatch/gigmatch/internal/middleware"
	"github.com/gigmatch/gigmatch/internal/service"
	"github.com/gigmatch/gigmatch/internal/util"
)

// DashboardHandler exposes the workbench endpoint.
type DashboardHandler struct {
	svc    *service.DashboardService
	logger *slog.Logger
}

// NewDashboardHandler builds a DashboardHandler.
func NewDashboardHandler(svc *service.DashboardService, logger *slog.Logger) *DashboardHandler {
	return &DashboardHandler{svc: svc, logger: logger}
}

// Get handles GET /dashboard.
func (h *DashboardHandler) Get(c *gin.Context) {
	u := middleware.GetCurrentUser(c)
	data, err := h.svc.Get(u.ID)
	if err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, data)
}
