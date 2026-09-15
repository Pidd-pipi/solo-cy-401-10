package constants

// User roles.
const (
	RoleRequester   = "requester"
	RoleFreelancer  = "freelancer"
	RoleBoth        = "both"
	RoleAdmin       = "admin"
)

// ValidRole reports whether a role is valid.
func ValidRole(s string) bool {
	switch s {
	case RoleRequester, RoleFreelancer, RoleBoth, RoleAdmin:
		return true
	}
	return false
}
