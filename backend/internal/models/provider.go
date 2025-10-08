package models

import (
	"time"
)

type Provider struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	Name          string    `json:"name" gorm:"not null"` // cloudflare, tencentcloud
	APIKey        string    `json:"-" gorm:"type:text"`   // encrypted
	APISecret     string    `json:"-" gorm:"type:text"`   // encrypted
	ExtraConfig   string    `json:"-" gorm:"type:text"`   // encrypted JSON for additional config
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	// Relations
	DNSRecords    []DNSRecord `json:"dns_records,omitempty" gorm:"foreignKey:ProviderID"`
}

// ProviderType constants
const (
	ProviderCloudflare   = "cloudflare"
	ProviderTencentCloud = "tencentcloud"
)
