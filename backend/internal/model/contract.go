package model

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// ContractStage is one payment milestone of a contract.
type ContractStage struct {
	Name      string `json:"name"`
	Amount    float64 `json:"amount"`
	Status    string `json:"status"` // pending / in_progress / done
	DueAt     string `json:"dueAt"`
}

// Contract is the signed agreement between requester and freelancer.
type Contract struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	ContractNo    string    `gorm:"size:64;uniqueIndex;not null" json:"contractNo"`
	TotalAmount   float64   `gorm:"type:decimal(14,2);not null" json:"totalAmount"`
	PaymentType   string    `gorm:"size:16;not null;default:one_time" json:"paymentType"`
	StagesJS      string    `gorm:"column:stages;type:text" json:"-"`
	Status        string    `gorm:"size:24;not null;default:pending_signature" json:"status"`
	RequirementID uint      `gorm:"index;not null" json:"requirementId"`
	PartyAID      uint      `gorm:"index;not null" json:"partyAId"`
	PartyBID      uint      `gorm:"index;not null" json:"partyBId"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"-"`

	// Computed fields.
	Stages    []ContractStage `gorm:"-" json:"stages"`
	PartyA    *User           `gorm:"foreignKey:PartyAID" json:"partyA"`
	PartyB    *User           `gorm:"foreignKey:PartyBID" json:"partyB"`
	Requirement *Requirement  `gorm:"foreignKey:RequirementID" json:"requirement"`
}

// BeforeSave serializes stages.
func (c *Contract) BeforeSave(_ *gorm.DB) error {
	if c.Stages != nil {
		raw, err := json.Marshal(c.Stages)
		if err != nil {
			return err
		}
		c.StagesJS = string(raw)
	}
	return nil
}

// AfterFind restores stages.
func (c *Contract) AfterFind(_ *gorm.DB) error {
	c.Stages = []ContractStage{}
	if c.StagesJS != "" {
		_ = json.Unmarshal([]byte(c.StagesJS), &c.Stages)
	}
	return nil
}
