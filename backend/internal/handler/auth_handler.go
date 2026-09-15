package handler

import (
	"log/slog"

	"github.com/gin-gonic/gin"

	"github.com/gigmatch/gigmatch/internal/dto"
	"github.com/gigmatch/gigmatch/internal/middleware"
	"github.com/gigmatch/gigmatch/internal/service"
	"github.com/gigmatch/gigmatch/internal/util"
)

// AuthHandler exposes auth endpoints.
type AuthHandler struct {
	svc    *service.AuthService
	logger *slog.Logger
}

// NewAuthHandler builds an AuthHandler.
func NewAuthHandler(svc *service.AuthService, logger *slog.Logger) *AuthHandler {
	return &AuthHandler{svc: svc, logger: logger}
}

// Register handles POST /auth/register.
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if !util.BindAndValidate(c, &req) {
		return
	}
	user, token, err := h.svc.Register(req)
	if err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, dto.AuthResponse{Token: token, User: user})
}

// Login handles POST /auth/login.
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if !util.BindAndValidate(c, &req) {
		return
	}
	user, token, err := h.svc.Login(req)
	if err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, dto.AuthResponse{Token: token, User: user})
}

// Me handles GET /auth/me.
func (h *AuthHandler) Me(c *gin.Context) {
	user, err := h.svc.GetByID(middleware.GetUserID(c))
	if err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, user)
}
