package handlers

import (
	"dnsmesh/internal/database"
	"dnsmesh/internal/models"
	"dnsmesh/internal/services"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// CreateRecordRequest represents the request to create a DNS record
type CreateRecordRequest struct {
	ProviderID   uint   `json:"provider_id" binding:"required"`
	ZoneID       string `json:"zone_id" binding:"required"`
	ZoneName     string `json:"zone_name" binding:"required"`
	FullDomain   string `json:"full_domain" binding:"required"`
	RecordType   string `json:"record_type" binding:"required"`
	TargetValue  string `json:"target_value" binding:"required"`
	TTL          int    `json:"ttl"`
	IsServer     bool   `json:"is_server"`
	ServerName   string `json:"server_name"`
	ServerRegion string `json:"server_region"`
	Notes        string `json:"notes"`
}

// ImportRecordsRequest represents batch import request
type ImportRecordsRequest struct {
	ProviderID uint               `json:"provider_id" binding:"required"`
	Records    []ImportRecordItem `json:"records" binding:"required"`
}

type ImportRecordItem struct {
	ZoneID           string `json:"zone_id"`
	ZoneName         string `json:"zone_name"`
	FullDomain       string `json:"full_domain"`
	RecordType       string `json:"record_type"`
	TargetValue      string `json:"target_value"`
	TTL              int    `json:"ttl"`
	ProviderRecordID string `json:"provider_record_id"`
	IsServer         bool   `json:"is_server"`
	ServerName       string `json:"server_name"`
	ServerRegion     string `json:"server_region"`
	Notes            string `json:"notes"`
}

const maxAuditRecordsPerProvider = 100

// GetRecords returns all DNS records grouped by server first, then unassigned by provider
func GetRecords(c *gin.Context) {
	var records []models.DNSRecord
	var providers []models.Provider

	// Fetch all records
	if err := database.DB.Where("managed = ?", true).Find(&records).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch records"})
		return
	}

	// Fetch all providers
	if err := database.DB.Find(&providers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch providers"})
		return
	}

	// Group records (server-first structure)
	grouped := services.GroupRecords(records, providers)

	c.JSON(http.StatusOK, grouped)
}

// findZoneForDomain finds the correct zone and provider for a given domain
func findZoneForDomain(domain string) (*models.DNSRecord, error) {
	// Try to find an existing record with a matching zone_name
	// Check from longest to shortest possible zone names
	parts := strings.Split(domain, ".")

	for i := 0; i < len(parts)-1; i++ {
		possibleZone := strings.Join(parts[i:], ".")

		var existingRecord models.DNSRecord
		err := database.DB.Where("zone_name = ?", possibleZone).First(&existingRecord).Error
		if err == nil {
			return &existingRecord, nil
		}
	}

	return nil, fmt.Errorf("no zone found for domain %s", domain)
}

// CreateRecord creates a new DNS record
func CreateRecord(c *gin.Context) {
	var req CreateRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate record type
	if req.RecordType != models.RecordTypeA && req.RecordType != models.RecordTypeCNAME {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid record type"})
		return
	}

	// Auto-detect zone and provider based on domain
	zoneRecord, err := findZoneForDomain(req.FullDomain)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot determine zone for domain: " + req.FullDomain + ". Please ensure the root domain exists in a provider."})
		return
	}

	// Use the detected provider and zone
	providerID := zoneRecord.ProviderID
	zoneID := zoneRecord.ZoneID
	zoneName := zoneRecord.ZoneName

	log.Printf("CreateRecord: Auto-detected zone '%s' (provider %d, zone %s) for domain '%s'",
		zoneName, providerID, zoneID, req.FullDomain)

	// Get provider
	var provider models.Provider
	if err := database.DB.First(&provider, providerID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Provider not found"})
		return
	}

	// Default TTL
	if req.TTL == 0 {
		req.TTL = 600
	}

	record := models.DNSRecord{
		ProviderID:   providerID,
		ZoneID:       zoneID,
		ZoneName:     zoneName,
		FullDomain:   req.FullDomain,
		RecordType:   req.RecordType,
		TargetValue:  req.TargetValue,
		TTL:          req.TTL,
		IsServer:     req.IsServer,
		ServerName:   req.ServerName,
		ServerRegion: req.ServerRegion,
		Notes:        req.Notes,
		Managed:      true,
	}

	// Create record on provider
	svc, err := getProviderService(&provider)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	providerRecordID, err := svc.CreateRecord(&record)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create record on provider: " + err.Error()})
		return
	}

	record.ProviderRecordID = providerRecordID

	// Save to database
	if err := database.DB.Create(&record).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save record"})
		return
	}

	// Log audit
	logAudit(c, models.ActionCreate, models.ResourceTypeRecord, record.ID, gin.H{
		"domain":      record.FullDomain,
		"record_type": record.RecordType,
		"target":      record.TargetValue,
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "Record created successfully",
		"record":  record,
	})
}

// UpdateRecord updates a DNS record
func UpdateRecord(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid record ID"})
		return
	}

	var record models.DNSRecord
	if err := database.DB.First(&record, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Record not found"})
		return
	}

	var req CreateRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if DNS-related fields have changed
	// DNS fields that need to be synced to provider: FullDomain, RecordType, TargetValue, TTL
	dnsFieldsChanged := record.FullDomain != req.FullDomain ||
		record.RecordType != req.RecordType ||
		record.TargetValue != req.TargetValue ||
		record.TTL != req.TTL

	// Update all fields (both DNS and local management fields)
	record.FullDomain = req.FullDomain
	record.RecordType = req.RecordType
	record.TargetValue = req.TargetValue
	record.TTL = req.TTL
	record.IsServer = req.IsServer
	record.ServerName = req.ServerName
	record.ServerRegion = req.ServerRegion
	record.Notes = req.Notes

	// Only call provider API if DNS fields changed
	if dnsFieldsChanged {
		log.Printf("UpdateRecord: DNS fields changed for record %d, updating provider", record.ID)

		// Get provider
		var provider models.Provider
		if err := database.DB.First(&provider, record.ProviderID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Provider not found"})
			return
		}

		// Update record on provider
		svc, err := getProviderService(&provider)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if err := svc.UpdateRecord(&record); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update record on provider: " + err.Error()})
			return
		}
	} else {
		log.Printf("UpdateRecord: Only local fields changed for record %d, skipping provider update", record.ID)
	}

	// Save to database (always save, whether DNS fields changed or not)
	if err := database.DB.Save(&record).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save record"})
		return
	}

	// Log audit
	auditDetails := gin.H{
		"domain":      record.FullDomain,
		"record_type": record.RecordType,
		"target":      record.TargetValue,
	}
	if !dnsFieldsChanged {
		auditDetails["local_only"] = true
	}
	logAudit(c, models.ActionUpdate, models.ResourceTypeRecord, record.ID, auditDetails)

	responseMessage := "Record updated successfully"
	if !dnsFieldsChanged {
		responseMessage = "Record metadata updated successfully (DNS unchanged)"
	}

	c.JSON(http.StatusOK, gin.H{
		"message": responseMessage,
		"record":  record,
	})
}

// HideRecord soft-deletes a DNS record by setting managed = false
// This removes the record from system control without deleting it from the DNS provider
// Can be used for both server records and regular records
func HideRecord(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid record ID"})
		return
	}

	var record models.DNSRecord
	if err := database.DB.First(&record, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Record not found"})
		return
	}

	// Soft delete: set managed = false
	record.Managed = false

	// Save to database
	if err := database.DB.Save(&record).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hide record"})
		return
	}

	// Log audit
	logAudit(c, models.ActionDelete, models.ResourceTypeRecord, record.ID, gin.H{
		"domain":      record.FullDomain,
		"record_type": record.RecordType,
		"action":      "hide",
	})

	c.JSON(http.StatusOK, gin.H{"message": "Record hidden from management"})
}

// DeleteRecord permanently deletes a DNS record from both provider and database
// Only allowed for non-server records (is_server = false)
func DeleteRecord(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid record ID"})
		return
	}

	var record models.DNSRecord
	if err := database.DB.First(&record, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Record not found"})
		return
	}

	// Prevent deletion of server records
	if record.IsServer {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete server records. Use hide instead."})
		return
	}

	// Get provider
	var provider models.Provider
	if err := database.DB.First(&provider, record.ProviderID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Provider not found"})
		return
	}

	// Delete from provider
	svc, err := getProviderService(&provider)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := svc.DeleteRecord(&record); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete record from provider: " + err.Error()})
		return
	}

	// Delete from database
	if err := database.DB.Delete(&record).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete record"})
		return
	}

	// Log audit
	logAudit(c, models.ActionDelete, models.ResourceTypeRecord, record.ID, gin.H{
		"domain":      record.FullDomain,
		"record_type": record.RecordType,
		"action":      "delete",
	})

	c.JSON(http.StatusOK, gin.H{"message": "Record deleted successfully"})
}

// ImportRecords batch imports DNS records
func ImportRecords(c *gin.Context) {
	var req ImportRecordsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("ImportRecords: Failed to bind JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("ImportRecords: Received request to import %d records for provider %d", len(req.Records), req.ProviderID)

	// Get provider
	var provider models.Provider
	if err := database.DB.First(&provider, req.ProviderID).Error; err != nil {
		log.Printf("ImportRecords: Provider %d not found: %v", req.ProviderID, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Provider not found"})
		return
	}

	var imported []models.DNSRecord
	var failed int

	for _, item := range req.Records {
		record := models.DNSRecord{
			ProviderID:       req.ProviderID,
			ZoneID:           item.ZoneID,
			ZoneName:         item.ZoneName,
			FullDomain:       item.FullDomain,
			RecordType:       item.RecordType,
			TargetValue:      item.TargetValue,
			TTL:              item.TTL,
			ProviderRecordID: item.ProviderRecordID,
			IsServer:         item.IsServer,
			ServerName:       item.ServerName,
			ServerRegion:     item.ServerRegion,
			Notes:            item.Notes,
			Managed:          true,
		}

		if err := database.DB.Create(&record).Error; err != nil {
			log.Printf("ImportRecords: Failed to import record %s: %v", item.FullDomain, err)
			failed++
			continue
		}

		imported = append(imported, record)
	}

	log.Printf("ImportRecords: Successfully imported %d records, failed %d", len(imported), failed)

	// Log audit
	logAudit(c, models.ActionCreate, models.ResourceTypeRecord, 0, gin.H{
		"action": "batch_import",
		"count":  len(imported),
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "Records imported successfully",
		"count":   len(imported),
		"records": imported,
	})
}

// ReanalyzeRecords re-syncs all providers and re-analyzes all records
func ReanalyzeRecords(c *gin.Context) {
	log.Println("ReanalyzeRecords: Starting re-analysis...")

	// Get all providers
	var providers []models.Provider
	if err := database.DB.Find(&providers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch providers"})
		return
	}

	type providerSyncSummary struct {
		ProviderID       uint                `json:"provider_id"`
		ProviderName     string              `json:"provider_name"`
		Synced           int                 `json:"synced"`
		Created          int                 `json:"created"`
		Updated          int                 `json:"updated"`
		Reimported       int                 `json:"reimported"`
		KeptHidden       int                 `json:"kept_hidden"`
		Records          []map[string]string `json:"records,omitempty"`
		RecordsTruncated bool                `json:"records_truncated,omitempty"`
		Errors           []string            `json:"errors,omitempty"`
	}

	providerStats := make(map[uint]*providerSyncSummary)

	// Collect all records from all providers with provider ID tracking
	type RecordWithProvider struct {
		ProviderID uint
		Record     services.DNSRecordSync
	}
	var allRecordsWithProvider []RecordWithProvider
	var allRecords []services.DNSRecordSync

	for _, provider := range providers {
		summary := &providerSyncSummary{ProviderID: provider.ID, ProviderName: provider.Name}
		providerStats[provider.ID] = summary

		svc, err := getProviderService(&provider)
		if err != nil {
			log.Printf("ReanalyzeRecords: Failed to get service for provider %d: %v", provider.ID, err)
			summary.Errors = append(summary.Errors, fmt.Sprintf("service_init: %v", err))
			continue
		}

		records, err := svc.SyncRecords()
		if err != nil {
			log.Printf("ReanalyzeRecords: Failed to sync provider %d: %v", provider.ID, err)
			summary.Errors = append(summary.Errors, fmt.Sprintf("sync: %v", err))
			continue
		}

		log.Printf("ReanalyzeRecords: Synced %d records from provider %d", len(records), provider.ID)

		// Track provider ID for each record
		for _, rec := range records {
			summary.Synced++
			if len(summary.Records) < maxAuditRecordsPerProvider {
				summary.Records = append(summary.Records, map[string]string{
					"zone":   rec.ZoneName,
					"domain": rec.FullDomain,
					"type":   rec.RecordType,
					"target": rec.TargetValue,
				})
			} else {
				summary.RecordsTruncated = true
			}
			allRecordsWithProvider = append(allRecordsWithProvider, RecordWithProvider{
				ProviderID: provider.ID,
				Record:     rec,
			})
		}

		allRecords = append(allRecords, records...)
	}

	log.Printf("ReanalyzeRecords: Total records to analyze: %d", len(allRecords))

	// Run analysis
	result := services.AnalyzeDNSRecords(allRecords)

	log.Printf("ReanalyzeRecords: Found %d server suggestions", len(result.ServerSuggestions))

	// First, sync all records to database (upsert)
	var synced int
	for _, rwp := range allRecordsWithProvider {
		summary, ok := providerStats[rwp.ProviderID]
		if !ok {
			summary = &providerSyncSummary{ProviderID: rwp.ProviderID}
			providerStats[rwp.ProviderID] = summary
		}
		var record models.DNSRecord

		// Try to find existing record
		err := database.DB.Where(
			"provider_id = ? AND zone_id = ? AND full_domain = ? AND record_type = ?",
			rwp.ProviderID, rwp.Record.ZoneID, rwp.Record.FullDomain, rwp.Record.RecordType,
		).First(&record).Error

		if err == nil {
			// Update existing record
			// Check if record was previously hidden (managed = false)
			wasHidden := !record.Managed
			contentChanged := record.TargetValue != rwp.Record.TargetValue ||
				record.TTL != rwp.Record.TTL ||
				record.ZoneName != rwp.Record.ZoneName

			record.TargetValue = rwp.Record.TargetValue
			record.TTL = rwp.Record.TTL
			record.ZoneName = rwp.Record.ZoneName
			record.ProviderRecordID = rwp.Record.ProviderRecordID

			// If was hidden and content hasn't changed, keep it hidden
			// If was hidden but content changed, re-import it (set managed = true)
			// If wasn't hidden, keep it as managed
			if wasHidden && !contentChanged {
				record.Managed = false
				log.Printf("ReanalyzeRecords: Keeping hidden record %s (no content change)", rwp.Record.FullDomain)
			} else {
				record.Managed = true
				if wasHidden && contentChanged {
					log.Printf("ReanalyzeRecords: Re-importing previously hidden record %s (content changed)", rwp.Record.FullDomain)
				}
			}

			if err := database.DB.Save(&record).Error; err != nil {
				log.Printf("ReanalyzeRecords: Failed to update record %s: %v", rwp.Record.FullDomain, err)
				summary.Errors = append(summary.Errors, fmt.Sprintf("update:%s: %v", rwp.Record.FullDomain, err))
				continue
			}

			if wasHidden && !contentChanged {
				summary.KeptHidden++
			} else {
				if contentChanged {
					summary.Updated++
				}
				if wasHidden && contentChanged {
					summary.Reimported++
				}
			}
		} else {
			// Create new record
			record = models.DNSRecord{
				ProviderID:       rwp.ProviderID,
				ZoneID:           rwp.Record.ZoneID,
				ZoneName:         rwp.Record.ZoneName,
				FullDomain:       rwp.Record.FullDomain,
				RecordType:       rwp.Record.RecordType,
				TargetValue:      rwp.Record.TargetValue,
				TTL:              rwp.Record.TTL,
				ProviderRecordID: rwp.Record.ProviderRecordID,
				Managed:          true,
				IsServer:         false,
			}

			if err := database.DB.Create(&record).Error; err != nil {
				log.Printf("ReanalyzeRecords: Failed to create record %s: %v", rwp.Record.FullDomain, err)
				summary.Errors = append(summary.Errors, fmt.Sprintf("create:%s: %v", rwp.Record.FullDomain, err))
				continue
			}

			summary.Created++
		}
		synced++
	}

	log.Printf("ReanalyzeRecords: Synced %d records to database", synced)

	// Then update server records based on suggestions
	var updated int
	for _, suggestion := range result.ServerSuggestions {
		// Find the record in database
		var record models.DNSRecord
		var err error

		// Try to find by domain first (more precise)
		if suggestion.Domain != "" {
			err = database.DB.Where("full_domain = ?", suggestion.Domain).First(&record).Error
		}

		// If not found by domain, try by IP
		if err != nil && suggestion.IP != "" {
			err = database.DB.Where("target_value = ?", suggestion.IP).First(&record).Error
		}

		if err != nil {
			log.Printf("ReanalyzeRecords: Record not found for suggestion %s (IP: %s): %v", suggestion.Domain, suggestion.IP, err)
			continue
		}

		// Update server fields
		record.IsServer = true
		record.ServerName = suggestion.SuggestedName
		record.ServerRegion = suggestion.SuggestedRegion

		if err := database.DB.Save(&record).Error; err != nil {
			log.Printf("ReanalyzeRecords: Failed to update record %d: %v", record.ID, err)
			continue
		}

		updated++
	}

	log.Printf("ReanalyzeRecords: Updated %d records as servers", updated)

	providerSummaries := make([]providerSyncSummary, 0, len(providers))
	for _, provider := range providers {
		if summary, ok := providerStats[provider.ID]; ok {
			providerSummaries = append(providerSummaries, *summary)
		}
	}

	// Log audit
	logAudit(c, models.ActionUpdate, models.ResourceTypeRecord, 0, gin.H{
		"action":         "reanalyze",
		"total_synced":   synced,
		"suggestions":    len(result.ServerSuggestions),
		"server_updates": updated,
		"providers":      providerSummaries,
	})

	c.JSON(http.StatusOK, gin.H{
		"message":      "Re-analysis completed",
		"suggestions":  len(result.ServerSuggestions),
		"updated":      updated,
		"providers":    providerSummaries,
		"total_synced": synced,
	})
}
