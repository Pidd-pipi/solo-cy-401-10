package service

import (
	"fmt"
	"log/slog"

	"github.com/gigmatch/gigmatch/internal/constants"
	"github.com/gigmatch/gigmatch/internal/dto"
	"github.com/gigmatch/gigmatch/internal/model"
	"github.com/gigmatch/gigmatch/internal/repository"
)

// RequirementService manages requirements.
type RequirementService struct {
	requirements *repository.RequirementRepository
	bids         *repository.BidRepository
	logs         *OperationLogService
	logger       *slog.Logger
}

// NewRequirementService builds a RequirementService.
func NewRequirementService(requirements *repository.RequirementRepository, bids *repository.BidRepository, logs *OperationLogService, logger *slog.Logger) *RequirementService {
	return &RequirementService{requirements: requirements, bids: bids, logs: logs, logger: logger}
}

// List returns requirements with filters and pagination.
func (s *RequirementService) List(status string, minBudget, maxBudget float64, skill string, page, pageSize int) ([]model.Requirement, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 12
	}
	items, total, err := s.requirements.List(status, minBudget, maxBudget, skill, page, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("list requirements: %w", err)
	}
	return items, total, nil
}

// Get loads a requirement detail.
func (s *RequirementService) Get(id uint) (*model.Requirement, error) {
	req, err := s.requirements.FindByID(id)
	if err != nil {
		return nil, err
	}
	return req, nil
}

// Create publishes a requirement.
func (s *RequirementService) Create(req dto.CreateRequirementRequest, userID uint, userName string, role string) (*model.Requirement, error) {
	if role == constants.RoleFreelancer {
		return nil, constants.NewAppError(constants.CodeForbidden, "自由职业者不能发布需求")
	}
	status := req.Status
	if status == "" {
		status = constants.RequirementOpen
	}
	if !constants.ValidRequirementStatus(status) {
		return nil, constants.NewAppError(constants.CodeBadRequest, "无效的需求状态")
	}
	skills := req.Skills
	if skills == nil {
		skills = []string{}
	}
	requirement := &model.Requirement{
		Title:       req.Title,
		Description: req.Description,
		MinBudget:   req.MinBudget,
		MaxBudget:   req.MaxBudget,
		Deadline:    dto.ParseDate(req.Deadline),
		Skills:      skills,
		Status:      status,
		PublisherID: userID,
	}
	if err := s.requirements.Create(requirement); err != nil {
		return nil, fmt.Errorf("create requirement: %w", err)
	}
	s.logs.Record(userID, userName, "requirement.create", "requirement", requirement.ID, fmt.Sprintf("发布需求 %s", requirement.Title))
	return requirement, nil
}

// Update edits a requirement owned by the caller.
func (s *RequirementService) Update(id uint, req dto.UpdateRequirementRequest, userID uint, userName string) (*model.Requirement, error) {
	r, err := s.requirements.FindByID(id)
	if err != nil {
		return nil, err
	}
	if r.PublisherID != userID {
		return nil, constants.ErrForbidden
	}
	if r.Status == constants.RequirementCompleted || r.Status == constants.RequirementCancelled {
		return nil, constants.NewAppError(constants.CodeConflict, "已完成或已取消的需求不可编辑")
	}
	r.Title = req.Title
	r.Description = req.Description
	r.MinBudget = req.MinBudget
	r.MaxBudget = req.MaxBudget
	r.Deadline = dto.ParseDate(req.Deadline)
	r.Skills = req.Skills
	if err := s.requirements.Update(r); err != nil {
		return nil, fmt.Errorf("update requirement: %w", err)
	}
	s.logs.Record(userID, userName, "requirement.update", "requirement", r.ID, "编辑需求")
	return r, nil
}

// UpdateStatus transitions a requirement's status.
func (s *RequirementService) UpdateStatus(id uint, status string, userID uint, userName string) (*model.Requirement, error) {
	if !constants.ValidRequirementStatus(status) {
		return nil, constants.NewAppError(constants.CodeBadRequest, "无效的需求状态")
	}
	r, err := s.requirements.FindByID(id)
	if err != nil {
		return nil, err
	}
	if r.PublisherID != userID {
		return nil, constants.ErrForbidden
	}
	r.Status = status
	if status == constants.RequirementCompleted && r.WinnerID > 0 {
		// mark the winner's contract context if any is handled by contract service
	}
	if err := s.requirements.Update(r); err != nil {
		return nil, fmt.Errorf("update requirement status: %w", err)
	}
	s.logs.Record(userID, userName, "requirement.status", "requirement", r.ID, fmt.Sprintf("需求状态 -> %s", status))
	return r, nil
}

// AcceptBid accepts a bid and creates the contract (delegated to contract service).
func (s *RequirementService) AcceptBid(requirementID, bidID, userID uint, userName string, paymentType string, contracts *ContractService) (*model.Contract, error) {
	r, err := s.requirements.FindByID(requirementID)
	if err != nil {
		return nil, err
	}
	if r.PublisherID != userID {
		return nil, constants.ErrForbidden
	}
	bid, err := s.bids.FindByID(bidID)
	if err != nil {
		return nil, err
	}
	if bid.RequirementID != requirementID {
		return nil, repository.ErrNotFound
	}
	if bid.Status != constants.BidPending {
		return nil, constants.NewAppError(constants.CodeConflict, "该报价已处理")
	}
	bid.Status = constants.BidAccepted
	if err := s.bids.Update(bid); err != nil {
		return nil, fmt.Errorf("accept bid: %w", err)
	}
	r.WinnerID = bid.BidderID
	r.Status = constants.RequirementInProgress
	if err := s.requirements.Update(r); err != nil {
		return nil, fmt.Errorf("update requirement after accept: %w", err)
	}
	contract, err := contracts.CreateFromBid(r, bid, userID, userName, paymentType)
	if err != nil {
		return nil, err
	}
	s.logs.Record(userID, userName, "requirement.accept_bid", "requirement", r.ID, fmt.Sprintf("采纳报价 %d", bidID))
	return contract, nil
}
