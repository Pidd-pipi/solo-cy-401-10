package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"gorm.io/gorm"

	"github.com/gigmatch/gigmatch/internal/constants"
	"github.com/gigmatch/gigmatch/internal/model"
	"github.com/gigmatch/gigmatch/internal/util"
)

// SeedService idempotently inserts demo users and requirements on first boot.
type SeedService struct {
	db     *gorm.DB
	logger *slog.Logger
}

// NewSeedService builds a SeedService.
func NewSeedService(db *gorm.DB, logger *slog.Logger) *SeedService {
	return &SeedService{db: db, logger: logger}
}

// EnsureSeedData creates the demo dataset if the users table is empty.
func (s *SeedService) EnsureSeedData(ctx context.Context) error {
	var count int64
	if err := s.db.WithContext(ctx).Model(&model.User{}).Count(&count).Error; err != nil {
		return fmt.Errorf("seed: count users: %w", err)
	}
	if count > 0 {
		return nil
	}

	hash, _ := util.HashPassword("demo123456")
	users := []*model.User{
		{Username: "requester01", PasswordHash: hash, Email: "requester01@example.com", Name: "张需求方", Role: constants.RoleRequester, Skills: []string{}, Bio: "互联网公司产品负责人", Rating: 4.8},
		{Username: "freelancer01", PasswordHash: hash, Email: "freelancer01@example.com", Name: "李自由职业者", Role: constants.RoleFreelancer, Skills: []string{"Go", "Gin", "MySQL"}, Bio: "全栈工程师，专注 Go 与云原生", Rating: 4.9},
		{Username: "freelancer02", PasswordHash: hash, Email: "freelancer02@example.com", Name: "王设计师", Role: constants.RoleFreelancer, Skills: []string{"UI设计", "Figma", "品牌"}, Bio: "资深 UI 设计师", Rating: 4.7},
		{Username: "admin", PasswordHash: hash, Email: "admin@example.com", Name: "平台管理员", Role: constants.RoleAdmin, Skills: []string{}, Bio: "平台运营", Rating: 5},
	}
	for _, u := range users {
		if err := s.db.WithContext(ctx).Create(u).Error; err != nil {
			return fmt.Errorf("seed: create user: %w", err)
		}
	}

	date := func(days int) time.Time {
		return time.Now().AddDate(0, 0, days)
	}

	requirements := []*model.Requirement{
		{Title: "开发企业官网后台管理系统", Description: "需要一个包含内容管理、用户权限、数据统计的企业官网后台，前后端分离，技术栈 Go + Vue。", MinBudget: 30000, MaxBudget: 60000, Deadline: date(30), Skills: []string{"Go", "Vue", "MySQL"}, Status: constants.RequirementOpen, PublisherID: users[0].ID},
		{Title: "设计品牌 VI 与落地页视觉", Description: "为新产品设计品牌 VI 体系和官网落地页视觉，需要提供完整设计规范与可交付源文件。", MinBudget: 15000, MaxBudget: 30000, Deadline: date(20), Skills: []string{"UI设计", "Figma", "品牌"}, Status: constants.RequirementOpen, PublisherID: users[0].ID},
		{Title: "开发数据可视化大屏", Description: "基于 ECharts 开发运营数据大屏，包含实时数据接入与多屏适配。", MinBudget: 20000, MaxBudget: 40000, Deadline: date(25), Skills: []string{"Vue", "ECharts", "WebSocket"}, Status: constants.RequirementBidding, PublisherID: users[0].ID},
	}
	for i := range requirements {
		if err := s.db.WithContext(ctx).Create(requirements[i]).Error; err != nil {
			return fmt.Errorf("seed: create requirement: %w", err)
		}
	}

	bids := []*model.Bid{
		{RequirementID: requirements[0].ID, BidderID: users[1].ID, Amount: 45000, DurationDays: 45, Proposal: "12 年 Go 全栈经验，可在一个月内交付 MVP 并上线。", Attachments: []string{"/uploads/portfolio-a.pdf"}, Status: constants.BidPending},
		{RequirementID: requirements[0].ID, BidderID: users[2].ID, Amount: 52000, DurationDays: 40, Proposal: "前后端一体完成，附带三个月免费维护。", Attachments: []string{}, Status: constants.BidPending},
		{RequirementID: requirements[1].ID, BidderID: users[2].ID, Amount: 22000, DurationDays: 15, Proposal: "包含 VI 基础系统与官网首页/内页视觉稿。", Attachments: []string{"/uploads/vi-case.pdf"}, Status: constants.BidPending},
	}
	for i := range bids {
		if err := s.db.WithContext(ctx).Create(bids[i]).Error; err != nil {
			return fmt.Errorf("seed: create bid: %w", err)
		}
	}

	s.logger.Info("seed data inserted",
		"users", len(users), "requirements", len(requirements), "bids", len(bids))
	return nil
}
