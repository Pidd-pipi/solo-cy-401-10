package service

import (
	"fmt"
	"log/slog"

	"github.com/gigmatch/gigmatch/internal/config"
	"github.com/gigmatch/gigmatch/internal/constants"
	"github.com/gigmatch/gigmatch/internal/dto"
	"github.com/gigmatch/gigmatch/internal/model"
	"github.com/gigmatch/gigmatch/internal/repository"
	"github.com/gigmatch/gigmatch/internal/util"
)

// AuthService handles registration and login.
type AuthService struct {
	cfg    *config.Config
	users  *repository.UserRepository
	logs   *OperationLogService
	logger *slog.Logger
}

// NewAuthService builds an AuthService.
func NewAuthService(cfg *config.Config, users *repository.UserRepository, logs *OperationLogService, logger *slog.Logger) *AuthService {
	return &AuthService{cfg: cfg, users: users, logs: logs, logger: logger}
}

// Register creates a new account and returns a token.
func (s *AuthService) Register(req dto.RegisterRequest) (*model.User, string, error) {
	exists, err := s.users.ExistsByUsername(req.Username)
	if err != nil {
		return nil, "", fmt.Errorf("register: %w", err)
	}
	if exists {
		return nil, "", constants.ErrConflict
	}
	hash, err := util.HashPassword(req.Password)
	if err != nil {
		return nil, "", fmt.Errorf("register: hash password: %w", err)
	}
	name := req.Name
	if name == "" {
		name = req.Username
	}
	user := &model.User{
		Username:     req.Username,
		PasswordHash: hash,
		Email:        req.Email,
		Name:         name,
		Role:         req.Role,
		Skills:       []string{},
	}
	if err := s.users.Create(user); err != nil {
		return nil, "", fmt.Errorf("register: %w", err)
	}
	token, err := util.GenerateToken(s.cfg.JWTSecret, user.ID, user.Role, s.cfg.JWTExpireDuration())
	if err != nil {
		return nil, "", fmt.Errorf("register: generate token: %w", err)
	}
	s.logs.Record(user.ID, user.Username, "auth.register", "user", user.ID, "注册账号")
	return user, token, nil
}

// Login authenticates a user and returns a token.
func (s *AuthService) Login(req dto.LoginRequest) (*model.User, string, error) {
	user, err := s.users.FindByUsername(req.Username)
	if err != nil {
		return nil, "", constants.NewAppError(constants.CodeUnauthorized, "用户名或密码错误")
	}
	if !util.CheckPassword(user.PasswordHash, req.Password) {
		return nil, "", constants.NewAppError(constants.CodeUnauthorized, "用户名或密码错误")
	}
	token, err := util.GenerateToken(s.cfg.JWTSecret, user.ID, user.Role, s.cfg.JWTExpireDuration())
	if err != nil {
		return nil, "", fmt.Errorf("login: generate token: %w", err)
	}
	s.logs.Record(user.ID, user.Username, "auth.login", "user", user.ID, "登录")
	return user, token, nil
}

// GetByID loads a user.
func (s *AuthService) GetByID(id uint) (*model.User, error) {
	return s.users.FindByID(id)
}
