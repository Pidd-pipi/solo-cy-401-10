package service

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	drivermysql "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/gigmatch/gigmatch/internal/constants"
	"github.com/gigmatch/gigmatch/internal/dto"
	"github.com/gigmatch/gigmatch/internal/model"
	"github.com/gigmatch/gigmatch/internal/repository"
)

// mysqlTestEnvVar points the accept-bid integration tests at a real MySQL
// server. It defaults to the database exposed by the project's docker compose
// stack (DB_PORT 33301, root credentials from .env.example).
const mysqlTestEnvVar = "GIGMATCH_TEST_MYSQL_DSN"

const defaultMysqlTestDSN = "root:root_pwd@tcp(127.0.0.1:33301)/?charset=utf8mb4&parseTime=True&loc=Local&timeout=3s"

// mysqlTestServerDSN returns the server-level DSN for the integration tests.
func mysqlTestServerDSN() string {
	if dsn := os.Getenv(mysqlTestEnvVar); dsn != "" {
		return dsn
	}
	return defaultMysqlTestDSN
}

var mysqlDBNameSanitizer = regexp.MustCompile(`[^0-9A-Za-z]+`)

// newMySQLAcceptBidDB builds a dedicated, freshly created database for one
// test case on a real MySQL server — every case starts from a clean state, so
// repeated runs of the whole group behave identically. It returns two
// independent handles: db runs the accept flow, verify is used only for
// read-back assertions against the persisted state.
func newMySQLAcceptBidDB(t *testing.T) (db *gorm.DB, verify *gorm.DB) {
	t.Helper()
	step := "连接真实 MySQL"
	serverDSN := mysqlTestServerDSN()
	admin, err := gorm.Open(mysql.Open(serverDSN), &gorm.Config{})
	if err != nil {
		t.Skipf("step=%s: 需要可用的 MySQL 服务器（可用 %s 指定 DSN）: %v", step, mysqlTestEnvVar, err)
	}
	step = "重建干净测试库"
	dbName := "gigmatch_it_" + mysqlDBNameSanitizer.ReplaceAllString(t.Name(), "_")
	if err := admin.Exec("DROP DATABASE IF EXISTS `" + dbName + "`").Error; err != nil {
		t.Fatalf("step=%s: drop database: %v", step, err)
	}
	if err := admin.Exec("CREATE DATABASE `" + dbName + "` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci").Error; err != nil {
		t.Fatalf("step=%s: create database: %v", step, err)
	}
	t.Cleanup(func() {
		_ = admin.Exec("DROP DATABASE IF EXISTS `" + dbName + "`").Error
	})

	cfg, err := drivermysql.ParseDSN(serverDSN)
	if err != nil {
		t.Fatalf("step=%s: parse dsn: %v", step, err)
	}
	cfg.DBName = dbName
	db, err = gorm.Open(mysql.Open(cfg.FormatDSN()), &gorm.Config{})
	if err != nil {
		t.Fatalf("step=%s: open test database: %v", step, err)
	}
	if err := db.AutoMigrate(&model.User{}, &model.Requirement{}, &model.Bid{}, &model.Contract{}, &model.OperationLog{}); err != nil {
		t.Fatalf("step=%s: auto migrate: %v", step, err)
	}
	// The pool must allow several truly independent connections; capping it at
	// one would silently serialize the concurrency these tests assert on.
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("step=%s: get sql db: %v", step, err)
	}
	sqlDB.SetMaxOpenConns(8)
	sqlDB.SetMaxIdleConns(8)

	verify, err = gorm.Open(mysql.Open(cfg.FormatDSN()), &gorm.Config{})
	if err != nil {
		t.Fatalf("step=%s: open verify handle: %v", step, err)
	}
	return db, verify
}

// mysqlAcceptBidServices wires the real services on top of db.
func mysqlAcceptBidServices(db *gorm.DB) (*RequirementService, *ContractService, *BidService) {
	logSvc := NewOperationLogService(repository.NewOperationLogRepository(db), discardLogger())
	reqSvc := NewRequirementService(repository.NewRequirementRepository(db), repository.NewBidRepository(db), logSvc, discardLogger())
	bidSvc := NewBidService(repository.NewBidRepository(db), repository.NewRequirementRepository(db), logSvc, discardLogger())
	contractSvc := NewContractService(repository.NewContractRepository(db), logSvc, discardLogger())
	return reqSvc, contractSvc, bidSvc
}

func mysqlCreateUser(t *testing.T, db *gorm.DB, username, role string) *model.User {
	t.Helper()
	u := &model.User{Username: username, Name: "集成测试-" + username, Role: role}
	if err := db.Create(u).Error; err != nil {
		t.Fatalf("step=准备数据: create user %s: %v", username, err)
	}
	return u
}

func mysqlCreateRequirement(t *testing.T, reqSvc *RequirementService, requester *model.User) *model.Requirement {
	t.Helper()
	r, err := reqSvc.Create(dto.CreateRequirementRequest{
		Title: "开发官网后台", Description: "需要一个功能完整的后台管理系统", MinBudget: 30000, MaxBudget: 60000, Skills: []string{"Go"},
	}, requester.ID, requester.Name, requester.Role)
	if err != nil {
		t.Fatalf("step=准备数据: create requirement: %v", err)
	}
	return r
}

func mysqlCreateBid(t *testing.T, bidSvc *BidService, requirementID uint, amount float64, freelancer *model.User) *model.Bid {
	t.Helper()
	b, err := bidSvc.Create(dto.CreateBidRequest{
		RequirementID: requirementID, Amount: amount, DurationDays: 30, Proposal: "我有丰富的 Go 开发经验，可按时交付。",
	}, freelancer.ID, freelancer.Name, freelancer.Role)
	if err != nil {
		t.Fatalf("step=准备数据: create bid for freelancer %d: %v", freelancer.ID, err)
	}
	return b
}

// acceptAttempt is the outcome of one concurrent accept call.
type acceptAttempt struct {
	contract *model.Contract
	err      error
}

// runConcurrentAccepts releases n goroutines simultaneously (one per bid) and
// collects their results. It also reports the maximum number of accepts that
// were genuinely in flight at the same time, so the test can prove the
// attempts overlapped on independent connections instead of being serialized.
func runConcurrentAccepts(t *testing.T, bids []model.Bid, accept func(bidID uint) (*model.Contract, error)) (results []acceptAttempt, maxInFlight int) {
	t.Helper()
	var wg sync.WaitGroup
	start := make(chan struct{})
	var inFlight, maxSeen atomic.Int32
	results = make([]acceptAttempt, len(bids))
	for i, b := range bids {
		wg.Add(1)
		go func(i int, bidID uint) {
			defer wg.Done()
			<-start
			cur := inFlight.Add(1)
			for {
				m := maxSeen.Load()
				if cur <= m || maxSeen.CompareAndSwap(m, cur) {
					break
				}
			}
			defer inFlight.Add(-1)
			c, err := accept(bidID)
			results[i] = acceptAttempt{contract: c, err: err}
		}(i, b.ID)
	}
	close(start)
	wg.Wait()
	return results, int(maxSeen.Load())
}

// classifyAcceptAttempts asserts exactly one success and conflicts everywhere
// else, returning the winning index.
func classifyAcceptAttempts(t *testing.T, results []acceptAttempt) (winner int) {
	t.Helper()
	winner = -1
	for i, res := range results {
		if res.err == nil {
			if res.contract == nil {
				t.Fatalf("step=并发采纳: 第 %d 个请求成功但合同为 nil", i)
			}
			if winner != -1 {
				t.Fatalf("step=并发采纳: 第 %d 和第 %d 个请求都成功了, 期望最多一个成功", winner, i)
			}
			winner = i
			continue
		}
		var appErr *constants.AppError
		if !errors.As(res.err, &appErr) || appErr.Code != constants.CodeConflict {
			t.Fatalf("step=并发采纳: 第 %d 个失败请求 err=%v, 期望 CodeConflict 业务冲突", i, res.err)
		}
	}
	if winner == -1 {
		t.Fatal("step=并发采纳: 没有任何请求成功, 期望恰好一个成功")
	}
	return winner
}

// assertPersistedContract verifies the single persisted contract matches the
// winning bid, including the unchanged stage amount rules.
func assertPersistedContract(t *testing.T, verify *gorm.DB, requirement *model.Requirement, winningBid model.Bid, requester *model.User) model.Contract {
	t.Helper()
	step := "回读校验合同"
	var contracts []model.Contract
	if err := verify.Where("requirement_id = ?", requirement.ID).Find(&contracts).Error; err != nil {
		t.Fatalf("step=%s: 查询合同失败: %v", step, err)
	}
	if len(contracts) != 1 {
		t.Fatalf("step=%s: 合同数量=%d, 期望恰好 1 份", step, len(contracts))
	}
	c := contracts[0]
	if c.ContractNo != fmt.Sprintf("CY-%d-%d", requirement.ID, winningBid.ID) {
		t.Fatalf("step=%s: 合同号=%q, 期望 CY-%d-%d", step, c.ContractNo, requirement.ID, winningBid.ID)
	}
	if c.PartyAID != requester.ID || c.PartyBID != winningBid.BidderID {
		t.Fatalf("step=%s: 合同双方=%d/%d, 期望 %d/%d", step, c.PartyAID, c.PartyBID, requester.ID, winningBid.BidderID)
	}
	if c.TotalAmount != winningBid.Amount || c.PaymentType != "one_time" || c.Status != constants.ContractPendingSignature {
		t.Fatalf("step=%s: 合同金额/支付/状态=%v/%s/%s 不符合预期", step, c.TotalAmount, c.PaymentType, c.Status)
	}
	if len(c.Stages) != 3 {
		t.Fatalf("step=%s: 合同阶段数=%d, 期望 3", step, len(c.Stages))
	}
	wantAmounts := []float64{winningBid.Amount * 0.3, winningBid.Amount * 0.4, winningBid.Amount * 0.3}
	wantStatuses := []string{"done", "in_progress", "pending"}
	for i, stage := range c.Stages {
		if stage.Amount != wantAmounts[i] || stage.Status != wantStatuses[i] {
			t.Fatalf("step=%s: 阶段 %d 金额/状态=%v/%s, 期望 %v/%s", step, i, stage.Amount, stage.Status, wantAmounts[i], wantStatuses[i])
		}
	}
	return c
}

// Concurrent accepts of the SAME bid on a real MySQL server with independent
// connections: exactly one succeeds, the rest get a clear conflict, and the
// persisted winner, bid status and contract stay consistent on read-back.
func TestAcceptBidMySQLConcurrentSameBid(t *testing.T) {
	db, verify := newMySQLAcceptBidDB(t)
	reqSvc, contractSvc, bidSvc := mysqlAcceptBidServices(db)

	requester := mysqlCreateUser(t, db, "it-req-same", constants.RoleRequester)
	freelancer := mysqlCreateUser(t, db, "it-free-same", constants.RoleFreelancer)
	requirement := mysqlCreateRequirement(t, reqSvc, requester)
	bid := mysqlCreateBid(t, bidSvc, requirement.ID, 45000, freelancer)

	const contenders = 4
	bids := make([]model.Bid, contenders)
	for i := range bids {
		bids[i] = *bid
	}
	results, maxInFlight := runConcurrentAccepts(t, bids, func(bidID uint) (*model.Contract, error) {
		return reqSvc.AcceptBid(requirement.ID, bidID, requester.ID, requester.Name, "", contractSvc)
	})
	if maxInFlight < 2 {
		t.Fatalf("step=并发度校验: 最大在途采纳数=%d, 期望 >=2（并发被压成了顺序执行）", maxInFlight)
	}
	winner := classifyAcceptAttempts(t, results)

	step := "回读校验报价与需求"
	var gotBid model.Bid
	if err := verify.First(&gotBid, bid.ID).Error; err != nil {
		t.Fatalf("step=%s: 回读报价失败: %v", step, err)
	}
	if gotBid.Status != constants.BidAccepted {
		t.Fatalf("step=%s: 报价状态=%q, 期望 accepted", step, gotBid.Status)
	}
	var gotReq model.Requirement
	if err := verify.First(&gotReq, requirement.ID).Error; err != nil {
		t.Fatalf("step=%s: 回读需求失败: %v", step, err)
	}
	if gotReq.Status != constants.RequirementInProgress || gotReq.WinnerID != freelancer.ID {
		t.Fatalf("step=%s: 需求状态/中标人=%q/%d, 期望 in_progress/%d", step, gotReq.Status, gotReq.WinnerID, freelancer.ID)
	}

	persisted := assertPersistedContract(t, verify, requirement, *bid, requester)
	if persisted.ID != results[winner].contract.ID {
		t.Fatalf("step=回读校验合同: 持久化合同 ID=%d 与成功请求返回的 ID=%d 不一致", persisted.ID, results[winner].contract.ID)
	}
}

// Concurrent accepts of DIFFERENT bids of the same requirement on a real MySQL
// server: exactly one succeeds, every loser gets a clear conflict, losing bids
// stay pending, and the requirement persists exactly one winner and one
// contract.
func TestAcceptBidMySQLConcurrentDifferentBids(t *testing.T) {
	db, verify := newMySQLAcceptBidDB(t)
	reqSvc, contractSvc, bidSvc := mysqlAcceptBidServices(db)

	requester := mysqlCreateUser(t, db, "it-req-diff", constants.RoleRequester)
	requirement := mysqlCreateRequirement(t, reqSvc, requester)

	const contenders = 4
	bids := make([]model.Bid, contenders)
	for i := 0; i < contenders; i++ {
		freelancer := mysqlCreateUser(t, db, fmt.Sprintf("it-free-diff-%d", i), constants.RoleFreelancer)
		bids[i] = *mysqlCreateBid(t, bidSvc, requirement.ID, 40000+float64(i)*1000, freelancer)
	}

	results, maxInFlight := runConcurrentAccepts(t, bids, func(bidID uint) (*model.Contract, error) {
		return reqSvc.AcceptBid(requirement.ID, bidID, requester.ID, requester.Name, "", contractSvc)
	})
	if maxInFlight < 2 {
		t.Fatalf("step=并发度校验: 最大在途采纳数=%d, 期望 >=2（并发被压成了顺序执行）", maxInFlight)
	}
	winner := classifyAcceptAttempts(t, results)

	step := "回读校验报价与需求"
	var gotReq model.Requirement
	if err := verify.First(&gotReq, requirement.ID).Error; err != nil {
		t.Fatalf("step=%s: 回读需求失败: %v", step, err)
	}
	if gotReq.Status != constants.RequirementInProgress || gotReq.WinnerID != bids[winner].BidderID {
		t.Fatalf("step=%s: 需求状态/中标人=%q/%d, 期望 in_progress/%d", step, gotReq.Status, gotReq.WinnerID, bids[winner].BidderID)
	}
	for i, b := range bids {
		var gotBid model.Bid
		if err := verify.First(&gotBid, b.ID).Error; err != nil {
			t.Fatalf("step=%s: 回读报价 %d 失败: %v", step, b.ID, err)
		}
		want := constants.BidPending
		if i == winner {
			want = constants.BidAccepted
		}
		if gotBid.Status != want {
			t.Fatalf("step=%s: 报价 %d 状态=%q, 期望 %q", step, b.ID, gotBid.Status, want)
		}
	}

	assertPersistedContract(t, verify, requirement, bids[winner], requester)
}

// The group is repeatable: running the whole package again must observe the
// same clean-state behavior (each case rebuilds its own database).
func TestAcceptBidMySQLSequentialRepeatability(t *testing.T) {
	db, verify := newMySQLAcceptBidDB(t)
	reqSvc, contractSvc, bidSvc := mysqlAcceptBidServices(db)

	requester := mysqlCreateUser(t, db, "it-req-seq", constants.RoleRequester)
	freelancer := mysqlCreateUser(t, db, "it-free-seq", constants.RoleFreelancer)
	requirement := mysqlCreateRequirement(t, reqSvc, requester)
	bid := mysqlCreateBid(t, bidSvc, requirement.ID, 45000, freelancer)

	if _, err := reqSvc.AcceptBid(requirement.ID, bid.ID, requester.ID, requester.Name, "", contractSvc); err != nil {
		t.Fatalf("step=首次采纳: %v", err)
	}
	_, err := reqSvc.AcceptBid(requirement.ID, bid.ID, requester.ID, requester.Name, "", contractSvc)
	var appErr *constants.AppError
	if !errors.As(err, &appErr) || appErr.Code != constants.CodeConflict {
		t.Fatalf("step=重复采纳: err=%v, 期望 CodeConflict 业务冲突", err)
	}

	step := "回读校验"
	var contractCount int64
	if err := verify.Model(&model.Contract{}).Where("requirement_id = ?", requirement.ID).Count(&contractCount).Error; err != nil {
		t.Fatalf("step=%s: 统计合同失败: %v", step, err)
	}
	if contractCount != 1 {
		t.Fatalf("step=%s: 合同数量=%d, 期望 1", step, contractCount)
	}
	var gotReq model.Requirement
	if err := verify.First(&gotReq, requirement.ID).Error; err != nil {
		t.Fatalf("step=%s: 回读需求失败: %v", step, err)
	}
	if gotReq.WinnerID != freelancer.ID || gotReq.Status != constants.RequirementInProgress {
		t.Fatalf("step=%s: 需求中标人/状态=%d/%q, 期望 %d/in_progress", step, gotReq.WinnerID, gotReq.Status, freelancer.ID)
	}
	if !strings.HasPrefix(gotReq.Title, "开发官网后台") {
		t.Fatalf("step=%s: 需求标题回读异常: %q", step, gotReq.Title)
	}
}
