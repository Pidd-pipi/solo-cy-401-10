package repository

import (
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/gigmatch/gigmatch/internal/model"
)

// ContractRepository persists contracts.
type ContractRepository struct {
	db *gorm.DB
}

// NewContractRepository builds a ContractRepository.
func NewContractRepository(db *gorm.DB) *ContractRepository {
	return &ContractRepository{db: db}
}

// Create inserts a contract.
func (r *ContractRepository) Create(c *model.Contract) error {
	if err := r.db.Create(c).Error; err != nil {
		return fmt.Errorf("create contract: %w", err)
	}
	return nil
}

// ListByParty returns contracts where the user is either party.
func (r *ContractRepository) ListByParty(userID uint) ([]model.Contract, error) {
	var contracts []model.Contract
	if err := r.db.Where("party_a_id = ? OR party_b_id = ?", userID, userID).
		Preload("PartyA").
		Preload("PartyB").
		Preload("Requirement").
		Order("created_at DESC").
		Find(&contracts).Error; err != nil {
		return nil, fmt.Errorf("list contracts: %w", err)
	}
	return contracts, nil
}

// FindByID loads a contract by primary key.
func (r *ContractRepository) FindByID(id uint) (*model.Contract, error) {
	var c model.Contract
	err := r.db.Preload("PartyA").
		Preload("PartyB").
		Preload("Requirement").
		First(&c, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("find contract by id: %w", err)
	}
	return &c, nil
}

// Update persists contract changes.
func (r *ContractRepository) Update(c *model.Contract) error {
	if err := r.db.Save(c).Error; err != nil {
		return fmt.Errorf("update contract: %w", err)
	}
	return nil
}

// WithTx returns a copy of the repository whose writes run inside tx.
func (r *ContractRepository) WithTx(tx *gorm.DB) *ContractRepository {
	return &ContractRepository{db: tx}
}

// CountByParty returns the number of contracts involving a user.
func (r *ContractRepository) CountByParty(userID uint) (int64, error) {
	var count int64
	if err := r.db.Model(&model.Contract{}).
		Where("party_a_id = ? OR party_b_id = ?", userID, userID).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("count contracts: %w", err)
	}
	return count, nil
}
