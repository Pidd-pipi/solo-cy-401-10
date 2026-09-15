package service

import (
	"fmt"
	"log/slog"

	"github.com/gigmatch/gigmatch/internal/constants"
	"github.com/gigmatch/gigmatch/internal/dto"
	"github.com/gigmatch/gigmatch/internal/model"
	"github.com/gigmatch/gigmatch/internal/repository"
)

// BidService manages bids.
type BidService struct {
	bids         *repository.BidRepository
	requirements *repository.RequirementRepository
	logs         *OperationLogService
	logger       *slog.Logger
}

// NewBidService builds a BidService.
func NewBidService(bids *repository.BidRepository, requirements *repository.RequirementRepository, logs *OperationLogService, logger *slog.Logger) *BidService {
	return &BidService{bids: bids, requirements: requirements, logs: logs, logger: logger}
}

// ListByRequirement returns bids of a requirement.
func (s *BidService) ListByRequirement(requirementID uint) ([]model.Bid, error) {
	list, err := s.bids.ListByRequirement(requirementID)
	if err != nil {
		return nil, fmt.Errorf("list bids: %w", err)
	}
	return list, nil
}

// Create submits a bid (freelancer role).
func (s *BidService) Create(req dto.CreateBidRequest, userID uint, userName string, role string) (*model.Bid, error) {
	if role == constants.RoleRequester {
		return nil, constants.NewAppError(constants.CodeForbidden, "需求方不能提交报价")
	}
	r, err := s.requirements.FindByID(req.RequirementID)
	if err != nil {
		return nil, err
	}
	if r.PublisherID == userID {
		return nil, constants.NewAppError(constants.CodeForbidden, "不能对自己发布的需求报价")
	}
	if r.Status != constants.RequirementOpen && r.Status != constants.RequirementBidding {
		return nil, constants.NewAppError(constants.CodeConflict, "该需求当前不可报价")
	}
	attachments := req.Attachments
	if attachments == nil {
		attachments = []string{}
	}
	bid := &model.Bid{
		RequirementID: req.RequirementID,
		BidderID:      userID,
		Amount:        req.Amount,
		DurationDays:  req.DurationDays,
		Proposal:      req.Proposal,
		Attachments:   attachments,
		Status:        constants.BidPending,
	}
	if err := s.bids.Create(bid); err != nil {
		return nil, fmt.Errorf("create bid: %w", err)
	}
	s.logs.Record(userID, userName, "bid.create", "bid", bid.ID, fmt.Sprintf("提交报价 %.2f", bid.Amount))
	return bid, nil
}

// Withdraw withdraws a pending bid owned by the caller.
func (s *BidService) Withdraw(id uint, userID uint, userName string) (*model.Bid, error) {
	bid, err := s.bids.FindByID(id)
	if err != nil {
		return nil, err
	}
	if bid.BidderID != userID {
		return nil, constants.ErrForbidden
	}
	if bid.Status != constants.BidPending {
		return nil, constants.NewAppError(constants.CodeConflict, "仅待审报价可撤回")
	}
	bid.Status = constants.BidWithdrawn
	if err := s.bids.Update(bid); err != nil {
		return nil, fmt.Errorf("withdraw bid: %w", err)
	}
	s.logs.Record(userID, userName, "bid.withdraw", "bid", bid.ID, "撤回报价")
	return bid, nil
}
