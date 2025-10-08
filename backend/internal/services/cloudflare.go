package services

import (
	"context"
	"dnsmesh/internal/models"
	"fmt"

	"github.com/cloudflare/cloudflare-go"
)

// CloudflareService handles Cloudflare API operations
type CloudflareService struct {
	apiToken string
}

// NewCloudflareService creates a new Cloudflare service
func NewCloudflareService(apiToken, _ string) *CloudflareService {
	return &CloudflareService{
		apiToken: apiToken,
	}
}

// SyncRecords fetches all DNS records from Cloudflare
func (s *CloudflareService) SyncRecords() ([]DNSRecordSync, error) {
	api, err := cloudflare.NewWithAPIToken(s.apiToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create Cloudflare client: %w", err)
	}

	ctx := context.Background()
	var allRecords []DNSRecordSync

	// List all zones
	zones, err := api.ListZones(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list zones: %w", err)
	}

	// For each zone, fetch DNS records
	for _, zone := range zones {
		records, _, err := api.ListDNSRecords(ctx, cloudflare.ZoneIdentifier(zone.ID), cloudflare.ListDNSRecordsParams{})
		if err != nil {
			return nil, fmt.Errorf("failed to list DNS records for zone %s: %w", zone.Name, err)
		}

		for _, record := range records {
			// Only sync A and CNAME records
			if record.Type != models.RecordTypeA && record.Type != models.RecordTypeCNAME {
				continue
			}

			allRecords = append(allRecords, DNSRecordSync{
				ZoneID:           zone.ID,
				ZoneName:         zone.Name,
				FullDomain:       record.Name,
				RecordType:       record.Type,
				TargetValue:      record.Content,
				TTL:              record.TTL,
				ProviderRecordID: record.ID,
			})
		}
	}

	return allRecords, nil
}

// CreateRecord creates a DNS record in Cloudflare
func (s *CloudflareService) CreateRecord(record *models.DNSRecord) (string, error) {
	api, err := cloudflare.NewWithAPIToken(s.apiToken)
	if err != nil {
		return "", fmt.Errorf("failed to create Cloudflare client: %w", err)
	}

	ctx := context.Background()

	createParams := cloudflare.CreateDNSRecordParams{
		Type:    record.RecordType,
		Name:    record.FullDomain,
		Content: record.TargetValue,
		TTL:     record.TTL,
	}

	resp, err := api.CreateDNSRecord(ctx, cloudflare.ZoneIdentifier(record.ZoneID), createParams)
	if err != nil {
		return "", fmt.Errorf("failed to create DNS record: %w", err)
	}

	return resp.ID, nil
}

// UpdateRecord updates a DNS record in Cloudflare
func (s *CloudflareService) UpdateRecord(record *models.DNSRecord) error {
	api, err := cloudflare.NewWithAPIToken(s.apiToken)
	if err != nil {
		return fmt.Errorf("failed to create Cloudflare client: %w", err)
	}

	ctx := context.Background()

	updateParams := cloudflare.UpdateDNSRecordParams{
		ID:      record.ProviderRecordID,
		Type:    record.RecordType,
		Name:    record.FullDomain,
		Content: record.TargetValue,
		TTL:     record.TTL,
	}

	_, err = api.UpdateDNSRecord(ctx, cloudflare.ZoneIdentifier(record.ZoneID), updateParams)
	if err != nil {
		return fmt.Errorf("failed to update DNS record: %w", err)
	}

	return nil
}

// DeleteRecord deletes a DNS record from Cloudflare
func (s *CloudflareService) DeleteRecord(record *models.DNSRecord) error {
	api, err := cloudflare.NewWithAPIToken(s.apiToken)
	if err != nil {
		return fmt.Errorf("failed to create Cloudflare client: %w", err)
	}

	ctx := context.Background()

	err = api.DeleteDNSRecord(ctx, cloudflare.ZoneIdentifier(record.ZoneID), record.ProviderRecordID)
	if err != nil {
		return fmt.Errorf("failed to delete DNS record: %w", err)
	}

	return nil
}

// TestConnection tests the API connection
func (s *CloudflareService) TestConnection() error {
	api, err := cloudflare.NewWithAPIToken(s.apiToken)
	if err != nil {
		return fmt.Errorf("failed to create Cloudflare client: %w", err)
	}

	ctx := context.Background()
	_, err = api.ListZones(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to Cloudflare: %w", err)
	}

	return nil
}
