package dto

// GenerateContractRequest is the payload for creating a contract from a bid.
type GenerateContractRequest struct {
	BidID       uint   `json:"bidId" validate:"required"`
	PaymentType string `json:"paymentType" validate:"required,oneof=installments one_time"`
}
