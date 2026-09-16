package service

import (
	"errors"
	"fmt"
	"path/filepath"
	"sync"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/gigmatch/gigmatch/internal/constants"
	"github.com/gigmatch/gigmatch/internal/dto"
	"github.com/gigmatch/gigmatch/internal/model"
	"github.com/gigmatch/gigmatch/internal/repository"
)

// newAcceptBidTestDB builds a file-backed SQLite database. _txlock=immediate
// plus _busy_timeout make concurrent write transactions serialize the same way
// row locks do on MySQL, instead of failing with SQLITE_BUSY/SNAPSHOT.
func newAcceptBidTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := fmt.Sprintf("file:%s?_busy_timeout=10000&_txlock=immediate", filepath.Join(t.TempDir(), "accept_bid.db"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.User{}, &model.Requirement{}, &model.Bid{}, &model.Contract{}, &model.OperationLog{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func setupAcceptBidFlow(t *testing.T, db *gorm.DB) (*RequirementService, *ContractService, *BidService, *model.User, *model.User, *model.Requirement, *model.Bid) {
	t.Helper()
	logSvc := NewOperationLogService(repository.NewOperationLogRepository(db), discardLogger())
	reqSvc := NewRequirementService(repository.NewRequirementRepository(db), repository.NewBidRepository(db), logSvc, discardLogger())
	bidSvc := NewBidService(repository.NewBidRepository(db), repository.NewRequirementRepository(db), logSvc, discardLogger())
	contractSvc := NewContractService(repository.NewContractRepository(db), logSvc, discardLogger())

	userRepo := repository.NewUserRepository(db)
	requester := &model.User{Username: "req-tx", Name: "需求方", Role: constants.RoleRequester}
	freelancer := &model.User{Username: "free-tx", Name: "自由职业者", Role: constants.RoleFreelancer}
	if err := userRepo.Create(requester); err != nil {
		t.Fatal(err)
	}
	if err := userRepo.Create(freelancer); err != nil {
		t.Fatal(err)
	}

	requirement, err := reqSvc.Create(dto.CreateRequirementRequest{
		Title: "开发官网后台", Description: "需要一个功能完整的后台管理系统", MinBudget: 30000, MaxBudget: 60000, Skills: []string{"Go"},
	}, requester.ID, requester.Name, requester.Role)
	if err != nil {
		t.Fatalf("create requirement: %v", err)
	}
	bid, err := bidSvc.Create(dto.CreateBidRequest{
		RequirementID: requirement.ID, Amount: 45000, DurationDays: 30, Proposal: "我有丰富的 Go 开发经验，可按时交付。",
	}, freelancer.ID, freelancer.Name, freelancer.Role)
	if err != nil {
		t.Fatalf("create bid: %v", err)
	}
	return reqSvc, contractSvc, bidSvc, requester, freelancer, requirement, bid
}

// Two requests accepting the same bid concurrently: exactly one succeeds, the
// other gets a clear conflict, and no duplicate side effects are committed.
func TestAcceptBidConcurrentAcceptsSingleWinner(t *testing.T) {
	db := newAcceptBidTestDB(t)
	reqSvc, contractSvc, _, requester, freelancer, requirement, bid := setupAcceptBidFlow(t, db)

	const contenders = 4
	var wg sync.WaitGroup
	start := make(chan struct{})
	errs := make([]error, contenders)
	contracts := make([]*model.Contract, contenders)
	for i := 0; i < contenders; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			c, err := reqSvc.AcceptBid(requirement.ID, bid.ID, requester.ID, requester.Name, "installments", contractSvc)
			contracts[i] = c
			errs[i] = err
		}(i)
	}
	close(start)
	wg.Wait()

	var succeeded, conflicted int
	for i, err := range errs {
		if err == nil {
			succeeded++
			if contracts[i] == nil {
				t.Fatal("successful accept returned nil contract")
			}
			continue
		}
		var appErr *constants.AppError
		if !errors.As(err, &appErr) || appErr.Code != constants.CodeConflict {
			t.Fatalf("losing accept got %v, want AppError with CodeConflict", err)
		}
		conflicted++
	}
	if succeeded != 1 || conflicted != contenders-1 {
		t.Fatalf("succeeded=%d conflicted=%d, want 1 and %d", succeeded, conflicted, contenders-1)
	}

	// Exactly one contract exists for the requirement.
	var contractCount int64
	if err := db.Model(&model.Contract{}).Where("requirement_id = ?", requirement.ID).Count(&contractCount).Error; err != nil {
		t.Fatal(err)
	}
	if contractCount != 1 {
		t.Fatalf("contracts=%d, want 1", contractCount)
	}

	// The bid is accepted exactly once and the requirement moved on with the winner.
	var gotBid model.Bid
	if err := db.First(&gotBid, bid.ID).Error; err != nil {
		t.Fatal(err)
	}
	if gotBid.Status != constants.BidAccepted {
		t.Fatalf("bid status=%q, want accepted", gotBid.Status)
	}
	var gotReq model.Requirement
	if err := db.First(&gotReq, requirement.ID).Error; err != nil {
		t.Fatal(err)
	}
	if gotReq.Status != constants.RequirementInProgress || gotReq.WinnerID != freelancer.ID {
		t.Fatalf("requirement status=%q winner=%d", gotReq.Status, gotReq.WinnerID)
	}
}

// Different bids of the same requirement accepted concurrently: exactly one
// accept wins, every loser gets a clear conflict, and the requirement ends up
// with a single winner and a single contract.
func TestAcceptBidConcurrentDifferentBidsSingleWinner(t *testing.T) {
	db := newAcceptBidTestDB(t)
	reqSvc, contractSvc, bidSvc, requester, _, requirement, firstBid := setupAcceptBidFlow(t, db)

	userRepo := repository.NewUserRepository(db)
	bids := []model.Bid{*firstBid}
	for i := 1; i < 4; i++ {
		freelancer := &model.User{Username: fmt.Sprintf("free-tx-%d", i), Name: "自由职业者", Role: constants.RoleFreelancer}
		if err := userRepo.Create(freelancer); err != nil {
			t.Fatal(err)
		}
		bid, err := bidSvc.Create(dto.CreateBidRequest{
			RequirementID: requirement.ID, Amount: 40000 + float64(i)*1000, DurationDays: 30, Proposal: "我可以按时高质量交付这个需求。",
		}, freelancer.ID, freelancer.Name, freelancer.Role)
		if err != nil {
			t.Fatalf("create bid %d: %v", i, err)
		}
		bids = append(bids, *bid)
	}

	var wg sync.WaitGroup
	start := make(chan struct{})
	errs := make([]error, len(bids))
	for i, b := range bids {
		wg.Add(1)
		go func(i int, b model.Bid) {
			defer wg.Done()
			<-start
			_, errs[i] = reqSvc.AcceptBid(requirement.ID, b.ID, requester.ID, requester.Name, "", contractSvc)
		}(i, b)
	}
	close(start)
	wg.Wait()

	winner := -1
	for i, err := range errs {
		if err == nil {
			if winner != -1 {
				t.Fatalf("more than one accept succeeded: %d and %d", winner, i)
			}
			winner = i
			continue
		}
		var appErr *constants.AppError
		if !errors.As(err, &appErr) || appErr.Code != constants.CodeConflict {
			t.Fatalf("losing accept got %v, want AppError with CodeConflict", err)
		}
	}
	if winner == -1 {
		t.Fatal("no accept succeeded")
	}

	// Exactly one contract, belonging to the winning bid and its bidder.
	var contractCount int64
	if err := db.Model(&model.Contract{}).Where("requirement_id = ?", requirement.ID).Count(&contractCount).Error; err != nil {
		t.Fatal(err)
	}
	if contractCount != 1 {
		t.Fatalf("contracts=%d, want 1", contractCount)
	}
	var gotContract model.Contract
	if err := db.Where("requirement_id = ?", requirement.ID).First(&gotContract).Error; err != nil {
		t.Fatal(err)
	}
	if gotContract.ContractNo != fmt.Sprintf("CY-%d-%d", requirement.ID, bids[winner].ID) ||
		gotContract.PartyBID != bids[winner].BidderID ||
		gotContract.TotalAmount != bids[winner].Amount {
		t.Fatalf("contract %+v does not match winning bid %d", gotContract, bids[winner].ID)
	}

	// The requirement has exactly the winner; the winning bid is accepted and
	// every losing bid is still pending (its transition rolled back).
	var gotReq model.Requirement
	if err := db.First(&gotReq, requirement.ID).Error; err != nil {
		t.Fatal(err)
	}
	if gotReq.Status != constants.RequirementInProgress || gotReq.WinnerID != bids[winner].BidderID {
		t.Fatalf("requirement status=%q winner=%d, want in_progress/%d", gotReq.Status, gotReq.WinnerID, bids[winner].BidderID)
	}
	for i, b := range bids {
		var gotBid model.Bid
		if err := db.First(&gotBid, b.ID).Error; err != nil {
			t.Fatal(err)
		}
		want := constants.BidPending
		if i == winner {
			want = constants.BidAccepted
		}
		if gotBid.Status != want {
			t.Fatalf("bid %d status=%q, want %q", b.ID, gotBid.Status, want)
		}
	}
}

// Once an accept succeeded, a later accept of another pending bid on the same
// requirement conflicts and overwrites nothing.
func TestAcceptBidSecondBidConflictsAfterSuccess(t *testing.T) {
	db := newAcceptBidTestDB(t)
	reqSvc, contractSvc, bidSvc, requester, freelancer, requirement, firstBid := setupAcceptBidFlow(t, db)

	userRepo := repository.NewUserRepository(db)
	second := &model.User{Username: "free-tx-second", Name: "自由职业者乙", Role: constants.RoleFreelancer}
	if err := userRepo.Create(second); err != nil {
		t.Fatal(err)
	}
	secondBid, err := bidSvc.Create(dto.CreateBidRequest{
		RequirementID: requirement.ID, Amount: 42000, DurationDays: 25, Proposal: "我有丰富的 Go 开发经验，可按时交付。",
	}, second.ID, second.Name, second.Role)
	if err != nil {
		t.Fatalf("create second bid: %v", err)
	}

	if _, err := reqSvc.AcceptBid(requirement.ID, firstBid.ID, requester.ID, requester.Name, "", contractSvc); err != nil {
		t.Fatalf("first accept: %v", err)
	}
	_, err = reqSvc.AcceptBid(requirement.ID, secondBid.ID, requester.ID, requester.Name, "", contractSvc)
	var appErr *constants.AppError
	if !errors.As(err, &appErr) || appErr.Code != constants.CodeConflict {
		t.Fatalf("second accept got %v, want AppError with CodeConflict", err)
	}
	if appErr.Message != "该需求已采纳其他报价" {
		t.Fatalf("second accept message=%q, want 该需求已采纳其他报价", appErr.Message)
	}

	// The first accept stands: winner, contract and bid states are untouched.
	var gotReq model.Requirement
	if err := db.First(&gotReq, requirement.ID).Error; err != nil {
		t.Fatal(err)
	}
	if gotReq.WinnerID != freelancer.ID || gotReq.Status != constants.RequirementInProgress {
		t.Fatalf("requirement winner=%d status=%q, want %d/in_progress", gotReq.WinnerID, gotReq.Status, freelancer.ID)
	}
	var gotSecond model.Bid
	if err := db.First(&gotSecond, secondBid.ID).Error; err != nil {
		t.Fatal(err)
	}
	if gotSecond.Status != constants.BidPending {
		t.Fatalf("losing bid status=%q, want pending", gotSecond.Status)
	}
	var contractCount int64
	if err := db.Model(&model.Contract{}).Where("requirement_id = ?", requirement.ID).Count(&contractCount).Error; err != nil {
		t.Fatal(err)
	}
	if contractCount != 1 {
		t.Fatalf("contracts=%d, want 1", contractCount)
	}
}

// A second accept of an already-accepted bid is rejected with a conflict.
func TestAcceptBidAlreadyAcceptedConflicts(t *testing.T) {
	db := newAcceptBidTestDB(t)
	reqSvc, contractSvc, _, requester, _, requirement, bid := setupAcceptBidFlow(t, db)

	if _, err := reqSvc.AcceptBid(requirement.ID, bid.ID, requester.ID, requester.Name, "", contractSvc); err != nil {
		t.Fatalf("first accept: %v", err)
	}
	_, err := reqSvc.AcceptBid(requirement.ID, bid.ID, requester.ID, requester.Name, "", contractSvc)
	var appErr *constants.AppError
	if !errors.As(err, &appErr) || appErr.Code != constants.CodeConflict {
		t.Fatalf("second accept got %v, want AppError with CodeConflict", err)
	}
}

// When contract creation fails mid-flow, the whole accept rolls back: the bid
// stays pending, the requirement stays open without a winner, and no partial
// contract is left behind.
func TestAcceptBidRollsBackOnContractFailure(t *testing.T) {
	db := newAcceptBidTestDB(t)
	reqSvc, contractSvc, _, requester, freelancer, requirement, bid := setupAcceptBidFlow(t, db)

	// Pre-insert a contract with the number CreateFromBid will generate, so the
	// insert inside the accept flow violates the unique index.
	blocker := &model.Contract{
		ContractNo:    fmt.Sprintf("CY-%d-%d", requirement.ID, bid.ID),
		TotalAmount:   1,
		PaymentType:   "one_time",
		Status:        constants.ContractPendingSignature,
		RequirementID: requirement.ID,
		PartyAID:      requester.ID,
		PartyBID:      freelancer.ID,
	}
	if err := db.Create(blocker).Error; err != nil {
		t.Fatalf("insert blocker contract: %v", err)
	}

	if _, err := reqSvc.AcceptBid(requirement.ID, bid.ID, requester.ID, requester.Name, "", contractSvc); err == nil {
		t.Fatal("AcceptBid() error = nil, want failure from duplicate contract number")
	}

	var gotBid model.Bid
	if err := db.First(&gotBid, bid.ID).Error; err != nil {
		t.Fatal(err)
	}
	if gotBid.Status != constants.BidPending {
		t.Fatalf("bid status=%q after rollback, want pending", gotBid.Status)
	}
	var gotReq model.Requirement
	if err := db.First(&gotReq, requirement.ID).Error; err != nil {
		t.Fatal(err)
	}
	if gotReq.Status != constants.RequirementOpen || gotReq.WinnerID != 0 {
		t.Fatalf("requirement status=%q winner=%d after rollback, want open/0", gotReq.Status, gotReq.WinnerID)
	}
	var contractCount int64
	if err := db.Model(&model.Contract{}).Where("requirement_id = ?", requirement.ID).Count(&contractCount).Error; err != nil {
		t.Fatal(err)
	}
	if contractCount != 1 {
		t.Fatalf("contracts=%d after rollback, want only the pre-existing one", contractCount)
	}
}
