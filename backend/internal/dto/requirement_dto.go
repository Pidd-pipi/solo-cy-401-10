package dto

import "time"

// CreateRequirementRequest is the payload for publishing a requirement.
type CreateRequirementRequest struct {
	Title       string   `json:"title" validate:"required,min=2,max=128"`
	Description string   `json:"description" validate:"required,min=10"`
	MinBudget   float64  `json:"minBudget" validate:"required,min=0"`
	MaxBudget   float64  `json:"maxBudget" validate:"required,gtfield=MinBudget"`
	Deadline    string   `json:"deadline"`
	Skills      []string `json:"skills"`
	Status      string   `json:"status"`
}

// UpdateRequirementRequest is the payload for editing a requirement.
type UpdateRequirementRequest struct {
	Title       string   `json:"title" validate:"required,min=2,max=128"`
	Description string   `json:"description" validate:"required,min=10"`
	MinBudget   float64  `json:"minBudget" validate:"required,min=0"`
	MaxBudget   float64  `json:"maxBudget" validate:"required,gtfield=MinBudget"`
	Deadline    string   `json:"deadline"`
	Skills      []string `json:"skills"`
}

// UpdateRequirementStatusRequest transitions a requirement's status.
type UpdateRequirementStatusRequest struct {
	Status string `json:"status" validate:"required"`
}

// AcceptBidRequest selects a winning bid.
type AcceptBidRequest struct {
	BidID uint `json:"bidId" validate:"required"`
}

// ParseDate parses an ISO date string into a time.Time.
func ParseDate(s string) time.Time {
	if s == "" {
		return time.Now().AddDate(0, 1, 0)
	}
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t
	}
	return time.Now().AddDate(0, 1, 0)
}
