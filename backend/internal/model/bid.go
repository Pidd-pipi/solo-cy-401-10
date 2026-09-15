package model

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// Bid is a freelancer's quote for a requirement.
type Bid struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	RequirementID uint      `gorm:"index;not null" json:"requirementId"`
	BidderID      uint      `gorm:"index;not null" json:"bidderId"`
	Amount        float64   `gorm:"type:decimal(14,2);not null" json:"amount"`
	DurationDays  int       `gorm:"not null" json:"durationDays"`
	Proposal      string    `gorm:"type:text" json:"proposal"`
	AttachmentsJS string    `gorm:"column:attachments;type:text" json:"-"`
	Status        string    `gorm:"size:16;not null;default:pending" json:"status"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"-"`

	// Computed fields.
	Attachments []string `gorm:"-" json:"attachments"`
	Bidder      *User    `gorm:"foreignKey:BidderID" json:"bidder"`
}

// BeforeSave serializes attachments.
func (b *Bid) BeforeSave(_ *gorm.DB) error {
	if b.Attachments != nil {
		raw, err := json.Marshal(b.Attachments)
		if err != nil {
			return err
		}
		b.AttachmentsJS = string(raw)
	}
	return nil
}

// AfterFind restores attachments.
func (b *Bid) AfterFind(_ *gorm.DB) error {
	b.Attachments = []string{}
	if b.AttachmentsJS != "" {
		_ = json.Unmarshal([]byte(b.AttachmentsJS), &b.Attachments)
	}
	return nil
}
