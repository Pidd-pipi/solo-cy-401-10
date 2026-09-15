package model

import "time"

// OperationLog records key operations for traceability.
type OperationLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index" json:"userId"`
	UserName  string    `gorm:"size:64" json:"userName"`
	Action    string    `gorm:"size:64" json:"action"`
	Entity    string    `gorm:"size:64" json:"entity"`
	EntityID  uint      `json:"entityId"`
	Detail    string    `gorm:"size:512" json:"detail"`
	CreatedAt time.Time `json:"createdAt"`
}
