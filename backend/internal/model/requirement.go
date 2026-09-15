package model

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// Requirement is a job posting by a requester.
type Requirement struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Title       string    `gorm:"size:128;not null" json:"title"`
	Description string    `gorm:"type:text" json:"description"`
	MinBudget   float64   `gorm:"type:decimal(14,2);not null" json:"minBudget"`
	MaxBudget   float64   `gorm:"type:decimal(14,2);not null" json:"maxBudget"`
	Deadline    time.Time `json:"deadline"`
	SkillsJS    string    `gorm:"column:skills;type:text" json:"-"`
	Status      string    `gorm:"size:24;not null;default:open" json:"status"`
	PublisherID uint      `gorm:"index;not null" json:"publisherId"`
	WinnerID    uint      `gorm:"index" json:"winnerId"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"-"`

	// Computed fields.
	Skills    []string `gorm:"-" json:"skills"`
	Publisher *User    `gorm:"foreignKey:PublisherID" json:"publisher"`
	Bids      []Bid    `gorm:"foreignKey:RequirementID" json:"bids"`
}

// BeforeSave serializes skills.
func (r *Requirement) BeforeSave(_ *gorm.DB) error {
	if r.Skills != nil {
		b, err := json.Marshal(r.Skills)
		if err != nil {
			return err
		}
		r.SkillsJS = string(b)
	}
	return nil
}

// AfterFind restores skills.
func (r *Requirement) AfterFind(_ *gorm.DB) error {
	r.Skills = []string{}
	if r.SkillsJS != "" {
		_ = json.Unmarshal([]byte(r.SkillsJS), &r.Skills)
	}
	return nil
}
