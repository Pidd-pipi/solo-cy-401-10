package service

import (
	"fmt"
	"log/slog"

	"gorm.io/gorm"

	"github.com/gigmatch/gigmatch/internal/constants"
	"github.com/gigmatch/gigmatch/internal/model"
	"github.com/gigmatch/gigmatch/internal/repository"
)

// ContractService manages contracts.
type ContractService struct {
	contracts *repository.ContractRepository
	logs      *OperationLogService
	logger    *slog.Logger
}

// NewContractService builds a ContractService.
func NewContractService(contracts *repository.ContractRepository, logs *OperationLogService, logger *slog.Logger) *ContractService {
	return &ContractService{contracts: contracts, logs: logs, logger: logger}
}

// WithTx returns a copy of the service whose contract (and audit log) writes
// run inside tx, so they commit or roll back together with the caller's
// transaction.
func (s *ContractService) WithTx(tx *gorm.DB) *ContractService {
	return &ContractService{contracts: s.contracts.WithTx(tx), logs: s.logs.WithTx(tx), logger: s.logger}
}

// ListByParty returns contracts involving the caller.
func (s *ContractService) ListByParty(userID uint) ([]model.Contract, error) {
	list, err := s.contracts.ListByParty(userID)
	if err != nil {
		return nil, fmt.Errorf("list contracts: %w", err)
	}
	return list, nil
}

// Get loads a contract.
func (s *ContractService) Get(id uint) (*model.Contract, error) {
	return s.contracts.FindByID(id)
}

// CreateFromBid builds a contract from an accepted bid.
func (s *ContractService) CreateFromBid(r *model.Requirement, bid *model.Bid, requesterID uint, requesterName string, paymentType string) (*model.Contract, error) {
	if paymentType == "" {
		paymentType = "one_time"
	}
	stages := []model.ContractStage{
		{Name: "项目启动", Amount: bid.Amount * 0.3, Status: "done", DueAt: "签约后3日内"},
		{Name: "中期交付", Amount: bid.Amount * 0.4, Status: "in_progress", DueAt: "工期过半"},
		{Name: "验收结项", Amount: bid.Amount * 0.3, Status: "pending", DueAt: "验收通过后"},
	}
	contract := &model.Contract{
		ContractNo:    fmt.Sprintf("CY-%d-%d", r.ID, bid.ID),
		TotalAmount:   bid.Amount,
		PaymentType:   paymentType,
		Stages:        stages,
		Status:        constants.ContractPendingSignature,
		RequirementID: r.ID,
		PartyAID:      requesterID,
		PartyBID:      bid.BidderID,
	}
	if err := s.contracts.Create(contract); err != nil {
		return nil, fmt.Errorf("create contract: %w", err)
	}
	s.logs.Record(requesterID, requesterName, "contract.create", "contract", contract.ID, fmt.Sprintf("生成合同 %s", contract.ContractNo))
	return contract, nil
}

// Sign confirms a contract by either party.
func (s *ContractService) Sign(id uint, userID uint, userName string) (*model.Contract, error) {
	c, err := s.contracts.FindByID(id)
	if err != nil {
		return nil, err
	}
	if c.PartyAID != userID && c.PartyBID != userID {
		return nil, constants.ErrForbidden
	}
	if c.Status != constants.ContractPendingSignature {
		return nil, constants.NewAppError(constants.CodeConflict, "合同当前不可签署")
	}
	c.Status = constants.ContractInProgress
	if err := s.contracts.Update(c); err != nil {
		return nil, fmt.Errorf("sign contract: %w", err)
	}
	s.logs.Record(userID, userName, "contract.sign", "contract", c.ID, "签署确认合同")
	return c, nil
}

// Complete confirms completion (requester side).
func (s *ContractService) Complete(id uint, userID uint, userName string) (*model.Contract, error) {
	c, err := s.contracts.FindByID(id)
	if err != nil {
		return nil, err
	}
	if c.PartyAID != userID {
		return nil, constants.ErrForbidden
	}
	if c.Status != constants.ContractInProgress && c.Status != constants.ContractPendingReview {
		return nil, constants.NewAppError(constants.CodeConflict, "合同当前不可完成确认")
	}
	c.Status = constants.ContractCompleted
	for i := range c.Stages {
		c.Stages[i].Status = "done"
	}
	if err := s.contracts.Update(c); err != nil {
		return nil, fmt.Errorf("complete contract: %w", err)
	}
	s.logs.Record(userID, userName, "contract.complete", "contract", c.ID, "确认合同完成")
	return c, nil
}
