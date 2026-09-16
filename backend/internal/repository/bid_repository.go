package repository

import (
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/gigmatch/gigmatch/internal/model"
)

// BidRepository persists bids.
type BidRepository struct {
	db *gorm.DB
}

// NewBidRepository builds a BidRepository.
func NewBidRepository(db *gorm.DB) *BidRepository {
	return &BidRepository{db: db}
}

// Create inserts a bid.
func (r *BidRepository) Create(b *model.Bid) error {
	if err := r.db.Create(b).Error; err != nil {
		return fmt.Errorf("create bid: %w", err)
	}
	return nil
}

// ListByRequirement returns bids of a requirement with bidders.
func (r *BidRepository) ListByRequirement(requirementID uint) ([]model.Bid, error) {
	var bids []model.Bid
	if err := r.db.Where("requirement_id = ?", requirementID).
		Preload("Bidder").
		Order("amount ASC").
		Find(&bids).Error; err != nil {
		return nil, fmt.Errorf("list bids: %w", err)
	}
	return bids, nil
}

// FindByID loads a bid by primary key.
func (r *BidRepository) FindByID(id uint) (*model.Bid, error) {
	var b model.Bid
	if err := r.db.Preload("Bidder").First(&b, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("find bid by id: %w", err)
	}
	return &b, nil
}

// Update persists bid changes.
func (r *BidRepository) Update(b *model.Bid) error {
	if err := r.db.Save(b).Error; err != nil {
		return fmt.Errorf("update bid: %w", err)
	}
	return nil
}

// TransitionStatus atomically moves a bid from one status to another. The
// expected current status is part of the UPDATE condition, so a concurrent
// transition of the same bid matches no row and ErrConflict is returned —
// exactly one caller wins the race.
func (r *BidRepository) TransitionStatus(id uint, from, to string) error {
	res := r.db.Model(&model.Bid{}).
		Where("id = ? AND status = ?", id, from).
		Update("status", to)
	if res.Error != nil {
		return fmt.Errorf("transition bid status: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrConflict
	}
	return nil
}

// WithTx returns a copy of the repository whose writes run inside tx.
func (r *BidRepository) WithTx(tx *gorm.DB) *BidRepository {
	return &BidRepository{db: tx}
}

// ListByBidder returns bids submitted by a user.
func (r *BidRepository) ListByBidder(userID uint) ([]model.Bid, error) {
	var bids []model.Bid
	if err := r.db.Where("bidder_id = ?", userID).
		Preload("Bidder").
		Order("created_at DESC").
		Find(&bids).Error; err != nil {
		return nil, fmt.Errorf("list bids by bidder: %w", err)
	}
	return bids, nil
}

// CountByBidder returns the number of bids submitted by a user.
func (r *BidRepository) CountByBidder(userID uint) (int64, error) {
	var count int64
	if err := r.db.Model(&model.Bid{}).Where("bidder_id = ?", userID).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("count bids by bidder: %w", err)
	}
	return count, nil
}
