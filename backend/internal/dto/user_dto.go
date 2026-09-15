package dto

// UpdateProfileRequest is the payload for editing a profile.
type UpdateProfileRequest struct {
	Name     string   `json:"name" validate:"max=64"`
	Avatar   string   `json:"avatar" validate:"max=255"`
	Skills   []string `json:"skills"`
	Bio      string   `json:"bio" validate:"max=512"`
	Contact  string   `json:"contact" validate:"max=128"`
	Email    string   `json:"email" validate:"omitempty,email"`
}
