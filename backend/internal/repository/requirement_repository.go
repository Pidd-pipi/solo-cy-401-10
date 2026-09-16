package repository

import (
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/gigmatch/gigmatch/internal/model"
)

// RequirementRepository persists requirements.
type RequirementRepository struct {
	db *gorm.DB
}

// NewRequirementRepository builds a RequirementRepository.
func NewRequirementRepository(db *gorm.DB) *RequirementRepository {
	return &RequirementRepository{db: db}
}

// Create inserts a requirement.
func (r *RequirementRepository) Create(req *model.Requirement) error {
	if err := r.db.Create(req).Error; err != nil {
		return fmt.Errorf("create requirement: %w", err)
	}
	return nil
}

// List returns requirements with filters and pagination.
func (r *RequirementRepository) List(status string, minBudget, maxBudget float64, skill string, page, pageSize int) ([]model.Requirement, int64, error) {
	q := r.db.Model(&model.Requirement{}).Preload("Publisher")
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if minBudget > 0 {
		q = q.Where("max_budget >= ?", minBudget)
	}
	if maxBudget > 0 {
		q = q.Where("min_budget <= ?", maxBudget)
	}
	if skill != "" {
		q = q.Where("skills LIKE ?", "%"+skill+"%")
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count requirements: %w", err)
	}
	var items []model.Requirement
	if err := q.Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&items).Error; err != nil {
		return nil, 0, fmt.Errorf("list requirements: %w", err)
	}
	return items, total, nil
}

// FindByID loads a requirement with publisher and bids.
func (r *RequirementRepository) FindByID(id uint) (*model.Requirement, error) {
	var req model.Requirement
	err := r.db.Preload("Publisher").
		Preload("Bids.Bidder").
		First(&req, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("find requirement by id: %w", err)
	}
	return &req, nil
}

// Update persists requirement changes.
func (r *RequirementRepository) Update(req *model.Requirement) error {
	if err := r.db.Save(req).Error; err != nil {
		return fmt.Errorf("update requirement: %w", err)
	}
	return nil
}

// Transaction runs fn inside a single database transaction: every write
// issued through tx-bound repositories commits together, or rolls back
// together when fn returns an error.
func (r *RequirementRepository) Transaction(fn func(tx *gorm.DB) error) error {
	return r.db.Transaction(fn)
}

// WithTx returns a copy of the repository whose writes run inside tx.
func (r *RequirementRepository) WithTx(tx *gorm.DB) *RequirementRepository {
	return &RequirementRepository{db: tx}
}

// ListByPublisher returns requirements published by a user.
func (r *RequirementRepository) ListByPublisher(userID uint) ([]model.Requirement, error) {
	var items []model.Requirement
	if err := r.db.Where("publisher_id = ?", userID).
		Preload("Publisher").
		Preload("Bids").
		Order("created_at DESC").
		Find(&items).Error; err != nil {
		return nil, fmt.Errorf("list requirements by publisher: %w", err)
	}
	return items, nil
}

// CountByPublisher returns the number of requirements published by a user.
func (r *RequirementRepository) CountByPublisher(userID uint) (int64, error) {
	var count int64
	if err := r.db.Model(&model.Requirement{}).Where("publisher_id = ?", userID).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("count requirements by publisher: %w", err)
	}
	return count, nil
}
