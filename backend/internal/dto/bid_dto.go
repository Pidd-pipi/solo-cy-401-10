package dto

// CreateBidRequest is the payload for submitting a bid.
type CreateBidRequest struct {
	RequirementID uint     `json:"requirementId" validate:"required"`
	Amount        float64  `json:"amount" validate:"required,min=0"`
	DurationDays  int      `json:"durationDays" validate:"required,min=1"`
	Proposal      string   `json:"proposal" validate:"required,min=10"`
	Attachments   []string `json:"attachments"`
}
