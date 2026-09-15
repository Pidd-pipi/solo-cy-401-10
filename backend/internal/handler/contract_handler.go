package handler

import (
	"log/slog"

	"github.com/gin-gonic/gin"

	"github.com/gigmatch/gigmatch/internal/middleware"
	"github.com/gigmatch/gigmatch/internal/service"
	"github.com/gigmatch/gigmatch/internal/util"
)

// ContractHandler exposes contract endpoints.
type ContractHandler struct {
	svc    *service.ContractService
	logger *slog.Logger
}

// NewContractHandler builds a ContractHandler.
func NewContractHandler(svc *service.ContractService, logger *slog.Logger) *ContractHandler {
	return &ContractHandler{svc: svc, logger: logger}
}

// List handles GET /contracts.
func (h *ContractHandler) List(c *gin.Context) {
	u := middleware.GetCurrentUser(c)
	contracts, err := h.svc.ListByParty(u.ID)
	if err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, contracts)
}

// Get handles GET /contracts/:id.
func (h *ContractHandler) Get(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	contract, err := h.svc.Get(id)
	if err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, contract)
}

// Sign handles POST /contracts/:id/sign.
func (h *ContractHandler) Sign(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	u := middleware.GetCurrentUser(c)
	contract, err := h.svc.Sign(id, u.ID, u.Name)
	if err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, contract)
}

// Complete handles POST /contracts/:id/complete.
func (h *ContractHandler) Complete(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	u := middleware.GetCurrentUser(c)
	contract, err := h.svc.Complete(id, u.ID, u.Name)
	if err != nil {
		util.Fail(c, err)
		return
	}
	util.OK(c, contract)
}
