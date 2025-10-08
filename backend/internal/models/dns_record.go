package models

import (
	"time"
)

type DNSRecord struct {
	ID               uint      `json:"id" gorm:"primaryKey"`
	ProviderID       uint      `json:"provider_id" gorm:"not null;index"`
	ZoneID           string    `json:"zone_id" gorm:"index"`         // Provider's zone ID
	ZoneName         string    `json:"zone_name" gorm:"index"`       // e.g., example.com
	FullDomain       string    `json:"full_domain" gorm:"index"`     // e.g., app1.example.com
	RecordType       string    `json:"record_type" gorm:"not null"`  // A, CNAME
	TargetValue      string    `json:"target_value" gorm:"not null"` // IP or domain
	TTL              int       `json:"ttl" gorm:"default:600"`
	IsServer         bool      `json:"is_server" gorm:"default:false;index"`
	ServerName       string    `json:"server_name"`   // e.g., hk-01
	ServerRegion     string    `json:"server_region"` // e.g., 香港
	Notes            string    `json:"notes" gorm:"type:text"`
	Active           bool      `json:"active" gorm:"default:true;index"`
	ProviderRecordID string    `json:"provider_record_id"`          // Provider's record ID
	Managed          bool      `json:"managed" gorm:"default:true"` // Whether managed by this system
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`

	// Relations
	Provider Provider `json:"provider,omitempty" gorm:"foreignKey:ProviderID"`
}

// RecordType constants
const (
	RecordTypeA     = "A"
	RecordTypeCNAME = "CNAME"
)
