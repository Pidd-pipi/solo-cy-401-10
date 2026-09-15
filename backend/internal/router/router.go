// Package router wires up the Gin engine and all routes.
package router

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/gigmatch/gigmatch/internal/config"
	"github.com/gigmatch/gigmatch/internal/constants"
	"github.com/gigmatch/gigmatch/internal/docs"
	"github.com/gigmatch/gigmatch/internal/handler"
	"github.com/gigmatch/gigmatch/internal/middleware"
	"github.com/gigmatch/gigmatch/internal/repository"
	"github.com/gigmatch/gigmatch/internal/service"
	"github.com/gigmatch/gigmatch/internal/util"
)

// Handlers aggregates every HTTP handler for assembly.
type Handlers struct {
	Auth         *handler.AuthHandler
	User         *handler.UserHandler
	Requirement  *handler.RequirementHandler
	Bid          *handler.BidHandler
	Contract     *handler.ContractHandler
	Dashboard    *handler.DashboardHandler
	OperationLog *handler.OperationLogHandler
}

// New assembles the Gin engine and registers all routes.
func New(cfg *config.Config, logger *slog.Logger, h *Handlers, users *repository.UserRepository, logs *service.OperationLogService, db *gorm.DB) *gin.Engine {
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	engine := gin.New()
	engine.Use(middleware.ErrorHandler(logger), middleware.RequestLogger(logger))
	engine.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.AllowedOrigins(),
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Request-ID"},
		ExposeHeaders:    []string{"X-Request-ID"},
		AllowCredentials: false,
	}))

	liveness := func(c *gin.Context) {
		c.JSON(http.StatusOK, util.Response{Code: constants.CodeSuccess, Message: constants.MsgOK, Data: gin.H{"status": "ok", "service": "gigmatch"}})
	}
	engine.GET("/healthz", liveness)
	engine.GET("/api/healthz", liveness)
	engine.GET("/readyz", readiness(db, logger))
	engine.GET("/api/readyz", readiness(db, logger))

	engine.GET("/docs", docs.IndexHandler())
	engine.GET("/docs/openapi.json", docs.OpenAPIHandler())

	api := engine.Group("/api/v1")
	auth := api.Group("/auth", middleware.RateLimit(cfg.AuthRateLimit, logger))
	{
		auth.POST("/register", h.Auth.Register)
		auth.POST("/login", h.Auth.Login)
	}

	protected := api.Group("")
	protected.Use(middleware.RateLimit(cfg.APIRateLimit, logger))
	protected.Use(middleware.JWTAuth(cfg, users, logger))
	protected.Use(middleware.AuditLogger(logs, logger))
	{
		protected.GET("/auth/me", h.Auth.Me)
		protected.GET("/users/:id", h.User.Get)
		protected.PATCH("/users/:id", h.User.Update)
		protected.GET("/requirements", h.Requirement.List)
		protected.POST("/requirements", h.Requirement.Create)
		protected.GET("/requirements/:id", h.Requirement.Get)
		protected.PUT("/requirements/:id", h.Requirement.Update)
		protected.POST("/requirements/:id/status", h.Requirement.UpdateStatus)
		protected.POST("/requirements/:id/accept-bid", h.Requirement.AcceptBid)
		protected.GET("/bids", h.Bid.ListByRequirement)
		protected.POST("/bids", h.Bid.Create)
		protected.POST("/bids/:id/withdraw", h.Bid.Withdraw)
		protected.GET("/contracts", h.Contract.List)
		protected.GET("/contracts/:id", h.Contract.Get)
		protected.POST("/contracts/:id/sign", h.Contract.Sign)
		protected.POST("/contracts/:id/complete", h.Contract.Complete)
		protected.GET("/dashboard", h.Dashboard.Get)
		protected.GET("/operation-logs", h.OperationLog.List)
	}

	return engine
}

func readiness(db *gorm.DB, logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		sqlDB, err := db.DB()
		if err != nil {
			logger.Error("readiness: get sql db", "error", err)
			c.JSON(http.StatusServiceUnavailable, util.Response{Code: constants.CodeInternal, Message: "database unavailable", Data: nil})
			return
		}
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()
		if err := sqlDB.PingContext(ctx); err != nil {
			logger.Error("readiness: ping database", "error", err)
			c.JSON(http.StatusServiceUnavailable, util.Response{Code: constants.CodeInternal, Message: "database unavailable", Data: nil})
			return
		}
		c.JSON(http.StatusOK, util.Response{Code: constants.CodeSuccess, Message: constants.MsgOK, Data: gin.H{"status": "ok", "service": "gigmatch"}})
	}
}
