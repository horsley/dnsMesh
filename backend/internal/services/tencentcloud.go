package services

import (
	"dnsmesh/internal/models"
	"fmt"
	"strconv"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	dnspod "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/dnspod/v20210323"
)

// TencentCloudService handles Tencent Cloud DNSPod API operations
type TencentCloudService struct {
	secretID  string
	secretKey string
}

// NewTencentCloudService creates a new Tencent Cloud service
func NewTencentCloudService(secretID, secretKey string) *TencentCloudService {
	return &TencentCloudService{
		secretID:  secretID,
		secretKey: secretKey,
	}
}

// getClient creates a DNSPod client
func (s *TencentCloudService) getClient() (*dnspod.Client, error) {
	credential := common.NewCredential(s.secretID, s.secretKey)
	cpf := profile.NewClientProfile()
	client, err := dnspod.NewClient(credential, "", cpf)
	if err != nil {
		return nil, fmt.Errorf("failed to create DNSPod client: %w", err)
	}
	return client, nil
}

// SyncRecords fetches all DNS records from Tencent Cloud DNSPod
func (s *TencentCloudService) SyncRecords() ([]DNSRecordSync, error) {
	client, err := s.getClient()
	if err != nil {
		return nil, err
	}

	var allRecords []DNSRecordSync

	// List all domains
	domainListReq := dnspod.NewDescribeDomainListRequest()
	domainListResp, err := client.DescribeDomainList(domainListReq)
	if err != nil {
		return nil, fmt.Errorf("failed to list domains: %w", err)
	}

	// For each domain, fetch DNS records
	for _, domain := range domainListResp.Response.DomainList {
		domainName := *domain.Name

		// List records for this domain
		recordListReq := dnspod.NewDescribeRecordListRequest()
		recordListReq.Domain = common.StringPtr(domainName)
		recordListResp, err := client.DescribeRecordList(recordListReq)
		if err != nil {
			return nil, fmt.Errorf("failed to list records for domain %s: %w", domainName, err)
		}

		for _, record := range recordListResp.Response.RecordList {
			recordType := *record.Type

			// Only sync A and CNAME records
			if recordType != models.RecordTypeA && recordType != models.RecordTypeCNAME {
				continue
			}

			fullDomain := domainName
			if *record.Name != "@" {
				fullDomain = *record.Name + "." + domainName
			}

			ttl := 600
			if record.TTL != nil {
				ttl = int(*record.TTL)
			}

			allRecords = append(allRecords, DNSRecordSync{
				ZoneID:           strconv.FormatUint(*domain.DomainId, 10),
				ZoneName:         domainName,
				FullDomain:       fullDomain,
				RecordType:       recordType,
				TargetValue:      *record.Value,
				TTL:              ttl,
				ProviderRecordID: strconv.FormatUint(*record.RecordId, 10),
			})
		}
	}

	return allRecords, nil
}

// CreateRecord creates a DNS record in Tencent Cloud DNSPod
func (s *TencentCloudService) CreateRecord(record *models.DNSRecord) (string, error) {
	client, err := s.getClient()
	if err != nil {
		return "", err
	}

	req := dnspod.NewCreateRecordRequest()
	req.Domain = common.StringPtr(record.ZoneName)

	// Extract subdomain from full domain
	subdomain := extractSubdomain(record.FullDomain, record.ZoneName)
	req.SubDomain = common.StringPtr(subdomain)

	req.RecordType = common.StringPtr(record.RecordType)
	req.Value = common.StringPtr(record.TargetValue)
	req.RecordLine = common.StringPtr("默认")

	if record.TTL > 0 {
		req.TTL = common.Uint64Ptr(uint64(record.TTL))
	}

	resp, err := client.CreateRecord(req)
	if err != nil {
		return "", fmt.Errorf("failed to create DNS record: %w", err)
	}

	return strconv.FormatUint(*resp.Response.RecordId, 10), nil
}

// UpdateRecord updates a DNS record in Tencent Cloud DNSPod
func (s *TencentCloudService) UpdateRecord(record *models.DNSRecord) error {
	client, err := s.getClient()
	if err != nil {
		return err
	}

	recordID, err := strconv.ParseUint(record.ProviderRecordID, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid record ID: %w", err)
	}

	req := dnspod.NewModifyRecordRequest()
	req.Domain = common.StringPtr(record.ZoneName)
	req.RecordId = common.Uint64Ptr(recordID)

	subdomain := extractSubdomain(record.FullDomain, record.ZoneName)
	req.SubDomain = common.StringPtr(subdomain)

	req.RecordType = common.StringPtr(record.RecordType)
	req.Value = common.StringPtr(record.TargetValue)
	req.RecordLine = common.StringPtr("默认")

	if record.TTL > 0 {
		req.TTL = common.Uint64Ptr(uint64(record.TTL))
	}

	_, err = client.ModifyRecord(req)
	if err != nil {
		return fmt.Errorf("failed to update DNS record: %w", err)
	}

	return nil
}

// DeleteRecord deletes a DNS record from Tencent Cloud DNSPod
func (s *TencentCloudService) DeleteRecord(record *models.DNSRecord) error {
	client, err := s.getClient()
	if err != nil {
		return err
	}

	recordIDUint, err := strconv.ParseUint(record.ProviderRecordID, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid record ID: %w", err)
	}

	req := dnspod.NewDeleteRecordRequest()
	req.Domain = common.StringPtr(record.ZoneName)
	req.RecordId = common.Uint64Ptr(recordIDUint)

	_, err = client.DeleteRecord(req)
	if err != nil {
		return fmt.Errorf("failed to delete DNS record: %w", err)
	}

	return nil
}

// TestConnection tests the API connection
func (s *TencentCloudService) TestConnection() error {
	client, err := s.getClient()
	if err != nil {
		return err
	}

	req := dnspod.NewDescribeDomainListRequest()
	_, err = client.DescribeDomainList(req)
	if err != nil {
		return fmt.Errorf("failed to connect to Tencent Cloud: %w", err)
	}

	return nil
}

// extractSubdomain extracts subdomain from full domain
// e.g., "app.example.com" and "example.com" -> "app"
// "@" represents the root domain
func extractSubdomain(fullDomain, zoneName string) string {
	if fullDomain == zoneName {
		return "@"
	}

	// Remove zone name from the end
	if len(fullDomain) > len(zoneName) {
		subdomain := fullDomain[:len(fullDomain)-len(zoneName)-1]
		return subdomain
	}

	return "@"
}
