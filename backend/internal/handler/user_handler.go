package handler

import (
	"log/slog"

	"github.com/gin-gonic/gin"

	"github.com/gigmatch/gigmatch/internal/dto"
	"github.com/gigmatch/gigmatch/internal/middleware"
	"github.com/gigmatch/gigmatch/internal/service"
	"github.com/gigmatch/gigmatch/internal/util"
)

// UserHandler exposes user profile endpoints.
type UserHandler struct {
	svc    *service.UserService
	logger *slog.Logger
}

// NewUserHandler builds a UserHandler.
func NewUserHandler(svc *service.UserService, logger *slog.Logger) *UserHandler {
	return &UserHandler{svc: svc, logger: logger}
}

// Get handles GET /users/:id.
func (h *UserHandler) Get(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	u, err := h.svc.Get(id)
	if err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, u)
}

// Update handles PATCH /users/:id.
func (h *UserHandler) Update(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	actor := middleware.GetCurrentUser(c)
	if actor.ID != id && actor.Role != "admin" {
		util.Fail(c, forbiddenErr())
		return
	}
	var req dto.UpdateProfileRequest
	if !util.BindAndValidate(c, &req) {
		return
	}
	u, err := h.svc.Update(id, req, actor.ID, actor.Name)
	if err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, u)
}
