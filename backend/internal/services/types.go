package services

import (
	"errors"

	"dnsmesh/internal/models"
)

// ErrRecordStatusNotSupported indicates the provider cannot toggle record status
var ErrRecordStatusNotSupported = errors.New("provider does not support toggling record status")

// DNSRecordSync represents a DNS record fetched from provider
type DNSRecordSync struct {
	ZoneID           string `json:"zone_id"`
	ZoneName         string `json:"zone_name"`
	FullDomain       string `json:"full_domain"`
	RecordType       string `json:"record_type"`
	TargetValue      string `json:"target_value"`
	TTL              int    `json:"ttl"`
	Active           bool   `json:"active"`
	ProviderRecordID string `json:"provider_record_id"`
}

// ServerSuggestion represents a suggested server from analysis
type ServerSuggestion struct {
	Domain          string   `json:"domain"`
	IP              string   `json:"ip"`
	MatchReason     string   `json:"match_reason"`
	Confidence      string   `json:"confidence"` // high, medium, low
	ReferencedBy    []string `json:"referenced_by"`
	SameIPDomains   []string `json:"same_ip_domains"`
	SuggestedName   string   `json:"suggested_name"`
	SuggestedRegion string   `json:"suggested_region"`
}

// AnalysisResult represents the result of DNS record analysis
type AnalysisResult struct {
	Records           []DNSRecordSync    `json:"records"`
	ServerSuggestions []ServerSuggestion `json:"server_suggestions"`
}

// ProviderCapabilities describe optional abilities of a provider
type ProviderCapabilities struct {
	SupportsRecordStatusToggle bool `json:"supports_record_status_toggle"`
}

// DNSProvider interface for different DNS providers
type DNSProvider interface {
	SyncRecords() ([]DNSRecordSync, error)
	CreateRecord(record *models.DNSRecord) (string, error)
	UpdateRecord(record *models.DNSRecord) error
	DeleteRecord(record *models.DNSRecord) error
	SetRecordStatus(record *models.DNSRecord, enabled bool) error
	TestConnection() error
}
