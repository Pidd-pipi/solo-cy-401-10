// Package dto defines request/response payloads with validator rules.
package dto

import "github.com/gigmatch/gigmatch/internal/model"

// RegisterRequest is the registration payload.
type RegisterRequest struct {
	Username string `json:"username" validate:"required,min=2,max=64"`
	Password string `json:"password" validate:"required,min=6,max=128"`
	Email    string `json:"email" validate:"omitempty,email"`
	Name     string `json:"name" validate:"max=64"`
	Role     string `json:"role" validate:"required,oneof=requester freelancer both"`
}

// LoginRequest is the login payload.
type LoginRequest struct {
	Username string `json:"username" validate:"required,min=2,max=64"`
	Password string `json:"password" validate:"required,min=6,max=128"`
}

// AuthResponse carries the JWT token and the user.
type AuthResponse struct {
	Token string     `json:"token"`
	User  *model.User `json:"user"`
}
