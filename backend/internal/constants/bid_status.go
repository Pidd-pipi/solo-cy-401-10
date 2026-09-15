package constants

// Bid statuses.
const (
	BidPending   = "pending"
	BidAccepted  = "accepted"
	BidRejected  = "rejected"
	BidWithdrawn = "withdrawn"
)

// ValidBidStatus reports whether a bid status is valid.
func ValidBidStatus(s string) bool {
	switch s {
	case BidPending, BidAccepted, BidRejected, BidWithdrawn:
		return true
	}
	return false
}
