package service

import (
	"log/slog"

	"github.com/gigmatch/gigmatch/internal/model"
	"github.com/gigmatch/gigmatch/internal/repository"
)

// DashboardData is the role-aware workbench payload.
type DashboardData struct {
	MyRequirements []model.Requirement `json:"myRequirements"`
	MyBids         []model.Bid         `json:"myBids"`
	MyContracts    []model.Contract    `json:"myContracts"`
	Counts         map[string]int64    `json:"counts"`
}

// DashboardService aggregates workbench data.
type DashboardService struct {
	requirements *repository.RequirementRepository
	bids         *repository.BidRepository
	contracts    *repository.ContractRepository
	logger       *slog.Logger
}

// NewDashboardService builds a DashboardService.
func NewDashboardService(requirements *repository.RequirementRepository, bids *repository.BidRepository, contracts *repository.ContractRepository, logger *slog.Logger) *DashboardService {
	return &DashboardService{requirements: requirements, bids: bids, contracts: contracts, logger: logger}
}

// Get returns the workbench payload for a user.
func (s *DashboardService) Get(userID uint) (*DashboardData, error) {
	myRequirements, err := s.requirements.ListByPublisher(userID)
	if err != nil {
		return nil, err
	}
	myBids, err := s.bids.ListByBidder(userID)
	if err != nil {
		return nil, err
	}
	myContracts, err := s.contracts.ListByParty(userID)
	if err != nil {
		return nil, err
	}
	reqCount, err := s.requirements.CountByPublisher(userID)
	if err != nil {
		return nil, err
	}
	bidCount, err := s.bids.CountByBidder(userID)
	if err != nil {
		return nil, err
	}
	contractCount, err := s.contracts.CountByParty(userID)
	if err != nil {
		return nil, err
	}
	return &DashboardData{
		MyRequirements: myRequirements,
		MyBids:         myBids,
		MyContracts:    myContracts,
		Counts: map[string]int64{
			"requirements": reqCount,
			"bids":         bidCount,
			"contracts":    contractCount,
		},
	}, nil
}
