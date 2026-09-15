// Command server is the entrypoint of the freelance platform service.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/gigmatch/gigmatch/internal/config"
	"github.com/gigmatch/gigmatch/internal/handler"
	"github.com/gigmatch/gigmatch/internal/logger"
	"github.com/gigmatch/gigmatch/internal/model"
	"github.com/gigmatch/gigmatch/internal/repository"
	"github.com/gigmatch/gigmatch/internal/router"
	"github.com/gigmatch/gigmatch/internal/service"
)

func main() {
	logger := logger.New(os.Getenv("APP_ENV"))
	if err := run(logger); err != nil {
		logger.Error("server exited with error", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	db, err := openDB(cfg, logger)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}

	if err := db.AutoMigrate(
		&model.User{},
		&model.Requirement{},
		&model.Bid{},
		&model.Contract{},
		&model.OperationLog{},
	); err != nil {
		return fmt.Errorf("auto migrate: %w", err)
	}

	seedSvc := service.NewSeedService(db, logger)
	if err := seedSvc.EnsureSeedData(context.Background()); err != nil {
		return fmt.Errorf("seed data: %w", err)
	}

	h, users, logs := buildHandlers(cfg, db, logger)

	engine := router.New(cfg, logger, h, users, logs, db)

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.ServerPort),
		Handler:           engine,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("server listening", "port", cfg.ServerPort, "env", cfg.Env)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	select {
	case err := <-errCh:
		return err
	case sig := <-quit:
		logger.Info("shutting down", "signal", sig.String())
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown server: %w", err)
	}
	return nil
}

func openDB(cfg *config.Config, logger *slog.Logger) (*gorm.DB, error) {
	var db *gorm.DB
	var err error
	retries := cfg.DBConnectRetries
	if retries < 0 {
		retries = 0
	}
	for attempt := 0; ; attempt++ {
		db, err = gorm.Open(mysql.Open(cfg.DSN()), &gorm.Config{
			Logger: gormlogger.Default.LogMode(gormlogger.Warn),
		})
		if err == nil {
			sqlDB, pingErr := db.DB()
			if pingErr == nil {
				if pingErr = sqlDB.Ping(); pingErr == nil {
					break
				}
			}
			err = pingErr
		}
		if attempt >= retries {
			return nil, fmt.Errorf("connect mysql after %d retries: %w", retries, err)
		}
		logger.Warn("database not ready, retrying", "attempt", attempt+1, "error", err)
		time.Sleep(time.Duration(cfg.DBConnectRetryIntervalSec) * time.Second)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql db: %w", err)
	}
	sqlDB.SetMaxOpenConns(cfg.DBMaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.DBMaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.DBConnMaxLifetimeMin) * time.Minute)
	return db, nil
}

func buildHandlers(cfg *config.Config, db *gorm.DB, logger *slog.Logger) (*router.Handlers, *repository.UserRepository, *service.OperationLogService) {
	userRepo := repository.NewUserRepository(db)
	reqRepo := repository.NewRequirementRepository(db)
	bidRepo := repository.NewBidRepository(db)
	contractRepo := repository.NewContractRepository(db)
	logRepo := repository.NewOperationLogRepository(db)

	logSvc := service.NewOperationLogService(logRepo, logger)
	authSvc := service.NewAuthService(cfg, userRepo, logSvc, logger)
	userSvc := service.NewUserService(userRepo, logSvc, logger)
	contractSvc := service.NewContractService(contractRepo, logSvc, logger)
	bidSvc := service.NewBidService(bidRepo, reqRepo, logSvc, logger)
	reqSvc := service.NewRequirementService(reqRepo, bidRepo, logSvc, logger)
	dashboardSvc := service.NewDashboardService(reqRepo, bidRepo, contractRepo, logger)

	return &router.Handlers{
		Auth:         handler.NewAuthHandler(authSvc, logger),
		User:         handler.NewUserHandler(userSvc, logger),
		Requirement:  handler.NewRequirementHandler(reqSvc, contractSvc, logger),
		Bid:          handler.NewBidHandler(bidSvc, logger),
		Contract:     handler.NewContractHandler(contractSvc, logger),
		Dashboard:    handler.NewDashboardHandler(dashboardSvc, logger),
		OperationLog: handler.NewOperationLogHandler(logSvc, logger),
	}, userRepo, logSvc
}
