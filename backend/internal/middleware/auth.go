// Package middleware provides HTTP middlewares shared by the API.
package middleware

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/gigmatch/gigmatch/internal/config"
	"github.com/gigmatch/gigmatch/internal/constants"
	"github.com/gigmatch/gigmatch/internal/model"
	"github.com/gigmatch/gigmatch/internal/repository"
	"github.com/gigmatch/gigmatch/internal/util"
)

const (
	ctxKeyUser   = "auth_user"
	ctxKeyUserID = "auth_user_id"
)

// JWTAuth validates the Bearer token and injects the current user.
func JWTAuth(cfg *config.Config, users *repository.UserRepository, logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, util.Response{
				Code:    constants.CodeUnauthorized,
				Message: constants.MsgUnauthorized,
			})
			return
		}
		claims, err := util.ParseToken(cfg.JWTSecret, strings.TrimPrefix(header, "Bearer "))
		if err != nil {
			logger.Warn("invalid jwt", "error", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, util.Response{
				Code:    constants.CodeUnauthorized,
				Message: constants.MsgUnauthorized,
			})
			return
		}
		user, err := users.FindByID(claims.UserID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, util.Response{
				Code:    constants.CodeUnauthorized,
				Message: constants.MsgUnauthorized,
			})
			return
		}
		c.Set(ctxKeyUser, user)
		c.Set(ctxKeyUserID, user.ID)
		c.Next()
	}
}

// GetCurrentUser returns the authenticated user from the context.
func GetCurrentUser(c *gin.Context) *model.User {
	if v, ok := c.Get(ctxKeyUser); ok {
		if u, ok := v.(*model.User); ok {
			return u
		}
	}
	return &model.User{Name: "anonymous"}
}

// GetUserID returns the authenticated user id from the context.
func GetUserID(c *gin.Context) uint {
	if v, ok := c.Get(ctxKeyUserID); ok {
		if id, ok := v.(uint); ok {
			return id
		}
	}
	return 0
}
