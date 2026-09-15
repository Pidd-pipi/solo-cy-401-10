package constants

// Requirement lifecycle statuses.
const (
	RequirementDraft          = "draft"
	RequirementOpen           = "open"
	RequirementBidding        = "bidding"
	RequirementInProgress     = "in_progress"
	RequirementPendingReview  = "pending_review"
	RequirementCompleted      = "completed"
	RequirementCancelled      = "cancelled"
)

// ValidRequirementStatus reports whether a status is valid.
func ValidRequirementStatus(s string) bool {
	switch s {
	case RequirementDraft, RequirementOpen, RequirementBidding, RequirementInProgress, RequirementPendingReview, RequirementCompleted, RequirementCancelled:
		return true
	}
	return false
}
