package models

import (
	"time"
)

type AuditLog struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	Action       string    `json:"action" gorm:"not null;index"` // create, update, delete, sync
	ResourceType string    `json:"resource_type" gorm:"index"`   // record, provider
	ResourceID   uint      `json:"resource_id"`
	Details      string    `json:"details" gorm:"type:text"` // JSON details
	IPAddress    string    `json:"ip_address"`
	CreatedAt    time.Time `json:"created_at" gorm:"index"`
}

// Action constants
const (
	ActionCreate = "create"
	ActionUpdate = "update"
	ActionDelete = "delete"
	ActionSync   = "sync"
)

// ResourceType constants
const (
	ResourceTypeRecord   = "record"
	ResourceTypeProvider = "provider"
)
