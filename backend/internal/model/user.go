// Package model defines the GORM entities of the freelance platform.
package model

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// User is a platform account (requester / freelancer / both / admin).
type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Username     string    `gorm:"size:64;uniqueIndex;not null" json:"username"`
	PasswordHash string    `gorm:"size:255;not null" json:"-"`
	Email        string    `gorm:"size:128" json:"email"`
	Name         string    `gorm:"size:64" json:"name"`
	Role         string    `gorm:"size:16;not null" json:"role"`
	Avatar       string    `gorm:"size:255" json:"avatar"`
	SkillsJS     string    `gorm:"column:skills;type:text" json:"-"`
	Bio          string    `gorm:"size:512" json:"bio"`
	Contact      string    `gorm:"size:128" json:"contact"`
	Rating       float64   `gorm:"type:decimal(4,2);not null;default:0" json:"rating"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"-"`

	// Computed field.
	Skills []string `gorm:"-" json:"skills"`
}

// BeforeSave serializes the skills array.
func (u *User) BeforeSave(_ *gorm.DB) error {
	if u.Skills != nil {
		b, err := json.Marshal(u.Skills)
		if err != nil {
			return err
		}
		u.SkillsJS = string(b)
	}
	return nil
}

// AfterFind restores the skills array.
func (u *User) AfterFind(_ *gorm.DB) error {
	u.Skills = []string{}
	if u.SkillsJS != "" {
		_ = json.Unmarshal([]byte(u.SkillsJS), &u.Skills)
	}
	return nil
}
