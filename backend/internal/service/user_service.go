package service

import (
	"fmt"
	"log/slog"

	"github.com/gigmatch/gigmatch/internal/dto"
	"github.com/gigmatch/gigmatch/internal/model"
	"github.com/gigmatch/gigmatch/internal/repository"
)

// UserService manages user profiles.
type UserService struct {
	users  *repository.UserRepository
	logs   *OperationLogService
	logger *slog.Logger
}

// NewUserService builds a UserService.
func NewUserService(users *repository.UserRepository, logs *OperationLogService, logger *slog.Logger) *UserService {
	return &UserService{users: users, logs: logs, logger: logger}
}

// Get loads a user profile.
func (s *UserService) Get(id uint) (*model.User, error) {
	u, err := s.users.FindByID(id)
	if err != nil {
		return nil, err
	}
	return u, nil
}

// Update edits a user profile.
func (s *UserService) Update(id uint, req dto.UpdateProfileRequest, actorID uint, actorName string) (*model.User, error) {
	u, err := s.users.FindByID(id)
	if err != nil {
		return nil, err
	}
	if req.Name != "" {
		u.Name = req.Name
	}
	if req.Avatar != "" {
		u.Avatar = req.Avatar
	}
	if req.Skills != nil {
		u.Skills = req.Skills
	}
	if req.Bio != "" {
		u.Bio = req.Bio
	}
	if req.Contact != "" {
		u.Contact = req.Contact
	}
	if req.Email != "" {
		u.Email = req.Email
	}
	if err := s.users.Update(u); err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}
	s.logs.Record(actorID, actorName, "user.update", "user", id, "更新个人资料")
	return u, nil
}
