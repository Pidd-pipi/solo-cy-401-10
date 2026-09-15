package handler

import (
	"log/slog"

	"github.com/gin-gonic/gin"

	"github.com/gigmatch/gigmatch/internal/dto"
	"github.com/gigmatch/gigmatch/internal/middleware"
	"github.com/gigmatch/gigmatch/internal/service"
	"github.com/gigmatch/gigmatch/internal/util"
)

// BidHandler exposes bid endpoints.
type BidHandler struct {
	svc    *service.BidService
	logger *slog.Logger
}

// NewBidHandler builds a BidHandler.
func NewBidHandler(svc *service.BidService, logger *slog.Logger) *BidHandler {
	return &BidHandler{svc: svc, logger: logger}
}

// ListByRequirement handles GET /bids?requirementId=.
func (h *BidHandler) ListByRequirement(c *gin.Context) {
	requirementID, err := parseQueryUint(c, "requirementId")
	if err != nil {
		util.Fail(c, err)
		return
	}
	bids, err := h.svc.ListByRequirement(requirementID)
	if err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, bids)
}

// Create handles POST /bids.
func (h *BidHandler) Create(c *gin.Context) {
	var req dto.CreateBidRequest
	if !util.BindAndValidate(c, &req) {
		return
	}
	u := middleware.GetCurrentUser(c)
	bid, err := h.svc.Create(req, u.ID, u.Name, u.Role)
	if err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, bid)
}

// Withdraw handles POST /bids/:id/withdraw.
func (h *BidHandler) Withdraw(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	u := middleware.GetCurrentUser(c)
	bid, err := h.svc.Withdraw(id, u.ID, u.Name)
	if err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, bid)
}
