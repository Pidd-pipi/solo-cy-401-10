package constants

// Contract statuses.
const (
	ContractPendingSignature = "pending_signature"
	ContractInProgress       = "in_progress"
	ContractPendingReview    = "pending_review"
	ContractCompleted        = "completed"
	ContractTerminated       = "terminated"
)

// ValidContractStatus reports whether a status is valid.
func ValidContractStatus(s string) bool {
	switch s {
	case ContractPendingSignature, ContractInProgress, ContractPendingReview, ContractCompleted, ContractTerminated:
		return true
	}
	return false
}
