package service

import (
	"testing"

	"github.com/gigmatch/gigmatch/internal/constants"
	"github.com/gigmatch/gigmatch/internal/dto"
	"github.com/gigmatch/gigmatch/internal/model"
	"github.com/gigmatch/gigmatch/internal/repository"
)

func TestRequirementServiceRoleAndUpdate(t *testing.T) {
	db := newFlowTestDB(t)
	logSvc := NewOperationLogService(repository.NewOperationLogRepository(db), discardLogger())
	reqSvc := NewRequirementService(repository.NewRequirementRepository(db), repository.NewBidRepository(db), logSvc, discardLogger())

	requester := &model.User{Username: "req-role", Name: "需求方", Role: constants.RoleRequester}
	freelancer := &model.User{Username: "free-role", Name: "自由职业者", Role: constants.RoleFreelancer}
	repo := repository.NewUserRepository(db)
	if err := repo.Create(requester); err != nil {
		t.Fatal(err)
	}
	if err := repo.Create(freelancer); err != nil {
		t.Fatal(err)
	}

	t.Run("freelancer forbidden to publish", func(t *testing.T) {
		_, err := reqSvc.Create(dto.CreateRequirementRequest{
			Title: "测试需求", Description: "这是一个用于测试的需求描述", MinBudget: 1000, MaxBudget: 5000, Skills: []string{"Go"},
		}, freelancer.ID, freelancer.Name, freelancer.Role)
		if err == nil {
			t.Fatal("Create() error = nil, want forbidden")
		}
	})

	req, err := reqSvc.Create(dto.CreateRequirementRequest{
		Title: "测试需求", Description: "这是一个用于测试的需求描述", MinBudget: 1000, MaxBudget: 5000, Skills: []string{"Go"},
	}, requester.ID, requester.Name, requester.Role)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if req.Status != constants.RequirementOpen {
		t.Fatalf("status = %q, want open", req.Status)
	}

	t.Run("owner can update", func(t *testing.T) {
		updated, err := reqSvc.Update(req.ID, dto.UpdateRequirementRequest{
			Title: "更新需求", Description: "这是更新后的需求描述", MinBudget: 2000, MaxBudget: 6000, Skills: []string{"Vue"},
		}, requester.ID, requester.Name)
		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}
		if updated.Title != "更新需求" {
			t.Fatalf("title = %q, want 更新需求", updated.Title)
		}
	})

	t.Run("non owner forbidden", func(t *testing.T) {
		_, err := reqSvc.Update(req.ID, dto.UpdateRequirementRequest{
			Title: "越权", Description: "这是一个越权的需求描述", MinBudget: 2000, MaxBudget: 6000, Skills: []string{"Vue"},
		}, freelancer.ID, freelancer.Name)
		if err == nil {
			t.Fatal("Update() error = nil, want forbidden")
		}
	})
}

func TestBidServiceCreateAndWithdraw(t *testing.T) {
	db := newFlowTestDB(t)
	logSvc := NewOperationLogService(repository.NewOperationLogRepository(db), discardLogger())
	reqSvc := NewRequirementService(repository.NewRequirementRepository(db), repository.NewBidRepository(db), logSvc, discardLogger())
	bidSvc := NewBidService(repository.NewBidRepository(db), repository.NewRequirementRepository(db), logSvc, discardLogger())

	requester := &model.User{Username: "req-bid", Name: "需求方", Role: constants.RoleRequester}
	freelancer := &model.User{Username: "free-bid", Name: "自由职业者", Role: constants.RoleFreelancer}
	repo := repository.NewUserRepository(db)
	if err := repo.Create(requester); err != nil {
		t.Fatal(err)
	}
	if err := repo.Create(freelancer); err != nil {
		t.Fatal(err)
	}

	req, err := reqSvc.Create(dto.CreateRequirementRequest{
		Title: "测试需求", Description: "这是一个用于测试的需求描述", MinBudget: 1000, MaxBudget: 5000, Skills: []string{"Go"},
	}, requester.ID, requester.Name, requester.Role)
	if err != nil {
		t.Fatalf("create requirement: %v", err)
	}

	t.Run("requester forbidden to bid", func(t *testing.T) {
		_, err := bidSvc.Create(dto.CreateBidRequest{RequirementID: req.ID, Amount: 2000, DurationDays: 10, Proposal: "我可以完成这个需求"}, requester.ID, requester.Name, requester.Role)
		if err == nil {
			t.Fatal("Create() error = nil, want forbidden")
		}
	})

	bid, err := bidSvc.Create(dto.CreateBidRequest{RequirementID: req.ID, Amount: 2000, DurationDays: 10, Proposal: "我可以完成这个需求"}, freelancer.ID, freelancer.Name, freelancer.Role)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if bid.Status != constants.BidPending {
		t.Fatalf("status = %q, want pending", bid.Status)
	}

	withdrawn, err := bidSvc.Withdraw(bid.ID, freelancer.ID, freelancer.Name)
	if err != nil {
		t.Fatalf("Withdraw() error = %v", err)
	}
	if withdrawn.Status != constants.BidWithdrawn {
		t.Fatalf("status = %q, want withdrawn", withdrawn.Status)
	}
}
