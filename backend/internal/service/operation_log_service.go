// Package service implements business logic on top of repositories.
package service

import (
	"fmt"
	"log/slog"

	"gorm.io/gorm"

	"github.com/gigmatch/gigmatch/internal/model"
	"github.com/gigmatch/gigmatch/internal/repository"
)

// OperationLogService records and reads operation logs.
type OperationLogService struct {
	logs   *repository.OperationLogRepository
	logger *slog.Logger
}

// NewOperationLogService builds an OperationLogService.
func NewOperationLogService(logs *repository.OperationLogRepository, logger *slog.Logger) *OperationLogService {
	return &OperationLogService{logs: logs, logger: logger}
}

// WithTx returns a copy of the service whose log writes run inside tx.
// Recording stays best-effort: a failed write is logged and swallowed, so
// it can never roll back the surrounding transaction.
func (s *OperationLogService) WithTx(tx *gorm.DB) *OperationLogService {
	return &OperationLogService{logs: s.logs.WithTx(tx), logger: s.logger}
}

// Record writes an operation log entry.
func (s *OperationLogService) Record(userID uint, userName, action, entity string, entityID uint, detail string) {
	entry := &model.OperationLog{
		UserID:   userID,
		UserName: userName,
		Action:   action,
		Entity:   entity,
		EntityID: entityID,
		Detail:   detail,
	}
	if err := s.logs.Create(entry); err != nil {
		s.logger.Warn("record operation log failed", "error", err)
	}
}

// List returns recent operation logs.
func (s *OperationLogService) List(limit int) ([]model.OperationLog, error) {
	logs, err := s.logs.List(limit)
	if err != nil {
		return nil, fmt.Errorf("list operation logs: %w", err)
	}
	return logs, nil
}
