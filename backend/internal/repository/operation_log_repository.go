package repository

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/gigmatch/gigmatch/internal/model"
)

// OperationLogRepository persists operation logs.
type OperationLogRepository struct {
	db *gorm.DB
}

// NewOperationLogRepository builds an OperationLogRepository.
func NewOperationLogRepository(db *gorm.DB) *OperationLogRepository {
	return &OperationLogRepository{db: db}
}

// Create inserts an operation log.
func (r *OperationLogRepository) Create(l *model.OperationLog) error {
	if err := r.db.Create(l).Error; err != nil {
		return fmt.Errorf("create operation log: %w", err)
	}
	return nil
}

// WithTx returns a copy of the repository whose writes run inside tx.
func (r *OperationLogRepository) WithTx(tx *gorm.DB) *OperationLogRepository {
	return &OperationLogRepository{db: tx}
}

// List returns recent operation logs.
func (r *OperationLogRepository) List(limit int) ([]model.OperationLog, error) {
	if limit <= 0 || limit > 200 {
		limit = 100
	}
	var logs []model.OperationLog
	if err := r.db.Order("created_at DESC").Limit(limit).Find(&logs).Error; err != nil {
		return nil, fmt.Errorf("list operation logs: %w", err)
	}
	return logs, nil
}
