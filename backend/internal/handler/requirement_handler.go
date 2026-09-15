package handler

import (
	"log/slog"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/gigmatch/gigmatch/internal/dto"
	"github.com/gigmatch/gigmatch/internal/middleware"
	"github.com/gigmatch/gigmatch/internal/service"
	"github.com/gigmatch/gigmatch/internal/util"
)

// RequirementHandler exposes requirement endpoints.
type RequirementHandler struct {
	svc       *service.RequirementService
	contracts *service.ContractService
	logger    *slog.Logger
}

// NewRequirementHandler builds a RequirementHandler.
func NewRequirementHandler(svc *service.RequirementService, contracts *service.ContractService, logger *slog.Logger) *RequirementHandler {
	return &RequirementHandler{svc: svc, contracts: contracts, logger: logger}
}

// List handles GET /requirements.
func (h *RequirementHandler) List(c *gin.Context) {
	status := c.Query("status")
	skill := c.Query("skill")
	minBudget, _ := strconv.ParseFloat(c.DefaultQuery("minBudget", "0"), 64)
	maxBudget, _ := strconv.ParseFloat(c.DefaultQuery("maxBudget", "0"), 64)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "12"))
	items, total, err := h.svc.List(status, minBudget, maxBudget, skill, page, pageSize)
	if err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, gin.H{"items": items, "total": total, "page": page, "page_size": pageSize})
}

// Get handles GET /requirements/:id.
func (h *RequirementHandler) Get(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	r, err := h.svc.Get(id)
	if err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, r)
}

// Create handles POST /requirements.
func (h *RequirementHandler) Create(c *gin.Context) {
	var req dto.CreateRequirementRequest
	if !util.BindAndValidate(c, &req) {
		return
	}
	u := middleware.GetCurrentUser(c)
	r, err := h.svc.Create(req, u.ID, u.Name, u.Role)
	if err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, r)
}

// Update handles PUT /requirements/:id.
func (h *RequirementHandler) Update(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var req dto.UpdateRequirementRequest
	if !util.BindAndValidate(c, &req) {
		return
	}
	u := middleware.GetCurrentUser(c)
	r, err := h.svc.Update(id, req, u.ID, u.Name)
	if err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, r)
}

// UpdateStatus handles POST /requirements/:id/status.
func (h *RequirementHandler) UpdateStatus(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var req dto.UpdateRequirementStatusRequest
	if !util.BindAndValidate(c, &req) {
		return
	}
	u := middleware.GetCurrentUser(c)
	r, err := h.svc.UpdateStatus(id, req.Status, u.ID, u.Name)
	if err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, r)
}

// AcceptBid handles POST /requirements/:id/accept-bid.
func (h *RequirementHandler) AcceptBid(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var req dto.AcceptBidRequest
	if !util.BindAndValidate(c, &req) {
		return
	}
	u := middleware.GetCurrentUser(c)
	contract, err := h.svc.AcceptBid(id, req.BidID, u.ID, u.Name, "", h.contracts)
	if err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, contract)
}
