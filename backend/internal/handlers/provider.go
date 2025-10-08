package handlers

import (
	"dnsmesh/internal/database"
	"dnsmesh/internal/models"
	"dnsmesh/internal/services"
	"dnsmesh/pkg/crypto"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// CreateProviderRequest represents the request to create a provider
type CreateProviderRequest struct {
	Name        string                 `json:"name" binding:"required"`
	APIKey      string                 `json:"api_key"`
	APISecret   string                 `json:"api_secret"`
	ExtraConfig map[string]interface{} `json:"extra_config"`
}

// GetProviders returns all providers
func GetProviders(c *gin.Context) {
	var providers []models.Provider

	if err := database.DB.Find(&providers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch providers"})
		return
	}

	// Don't expose sensitive data
	for i := range providers {
		providers[i].APIKey = ""
		providers[i].APISecret = ""
		providers[i].ExtraConfig = ""
	}

	c.JSON(http.StatusOK, gin.H{"providers": providers})
}

// CreateProvider creates a new DNS provider
func CreateProvider(c *gin.Context) {
	var req CreateProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("CreateProvider: Failed to bind JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("CreateProvider: Received request for provider %s", req.Name)

	// Validate provider name
	if req.Name != models.ProviderCloudflare && req.Name != models.ProviderTencentCloud {
		log.Printf("CreateProvider: Invalid provider name: %s", req.Name)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid provider name"})
		return
	}

	// Encrypt credentials
	encryptedKey, err := crypto.Encrypt(req.APIKey)
	if err != nil {
		log.Printf("CreateProvider: Failed to encrypt API key: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encrypt API key"})
		return
	}

	var encryptedSecret string
	if req.APISecret != "" {
		encryptedSecret, err = crypto.Encrypt(req.APISecret)
		if err != nil {
			log.Printf("CreateProvider: Failed to encrypt API secret: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encrypt API secret"})
			return
		}
	}

	var encryptedExtra string
	if req.ExtraConfig != nil {
		extraJSON, _ := json.Marshal(req.ExtraConfig)
		encryptedExtra, err = crypto.Encrypt(string(extraJSON))
		if err != nil {
			log.Printf("CreateProvider: Failed to encrypt extra config: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encrypt extra config"})
			return
		}
	}

	provider := models.Provider{
		Name:        req.Name,
		APIKey:      encryptedKey,
		APISecret:   encryptedSecret,
		ExtraConfig: encryptedExtra,
	}

	// Test connection before saving
	log.Printf("CreateProvider: Testing connection for provider %s", req.Name)
	if err := testProviderConnection(&provider); err != nil {
		log.Printf("CreateProvider: Connection test failed: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to connect to provider: " + err.Error()})
		return
	}

	if err := database.DB.Create(&provider).Error; err != nil {
		log.Printf("CreateProvider: Failed to save provider to database: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create provider"})
		return
	}

	log.Printf("CreateProvider: Successfully created provider ID %d", provider.ID)

	// Log audit
	logAudit(c, models.ActionCreate, models.ResourceTypeProvider, provider.ID, gin.H{
		"provider_name": provider.Name,
	})

	provider.APIKey = ""
	provider.APISecret = ""
	provider.ExtraConfig = ""

	c.JSON(http.StatusOK, gin.H{
		"message":  "Provider created successfully",
		"provider": provider,
	})
}

// UpdateProvider updates a provider
func UpdateProvider(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid provider ID"})
		return
	}

	var provider models.Provider
	if err := database.DB.First(&provider, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Provider not found"})
		return
	}

	var req CreateProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update credentials if provided
	if req.APIKey != "" {
		encryptedKey, err := crypto.Encrypt(req.APIKey)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encrypt API key"})
			return
		}
		provider.APIKey = encryptedKey
	}

	if req.APISecret != "" {
		encryptedSecret, err := crypto.Encrypt(req.APISecret)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encrypt API secret"})
			return
		}
		provider.APISecret = encryptedSecret
	}

	if req.ExtraConfig != nil {
		extraJSON, _ := json.Marshal(req.ExtraConfig)
		encryptedExtra, err := crypto.Encrypt(string(extraJSON))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encrypt extra config"})
			return
		}
		provider.ExtraConfig = encryptedExtra
	}

	// Test connection
	if err := testProviderConnection(&provider); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to connect to provider: " + err.Error()})
		return
	}

	if err := database.DB.Save(&provider).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update provider"})
		return
	}

	// Log audit
	logAudit(c, models.ActionUpdate, models.ResourceTypeProvider, provider.ID, gin.H{
		"provider_name": provider.Name,
	})

	provider.APIKey = ""
	provider.APISecret = ""
	provider.ExtraConfig = ""

	c.JSON(http.StatusOK, gin.H{
		"message":  "Provider updated successfully",
		"provider": provider,
	})
}

// DeleteProvider deletes a provider
func DeleteProvider(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid provider ID"})
		return
	}

	var provider models.Provider
	if err := database.DB.First(&provider, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Provider not found"})
		return
	}

	// Delete all associated records first
	if err := database.DB.Where("provider_id = ?", id).Delete(&models.DNSRecord{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete associated records"})
		return
	}

	if err := database.DB.Delete(&provider).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete provider"})
		return
	}

	// Log audit
	logAudit(c, models.ActionDelete, models.ResourceTypeProvider, provider.ID, gin.H{
		"provider_name": provider.Name,
	})

	c.JSON(http.StatusOK, gin.H{"message": "Provider deleted successfully"})
}

// SyncProvider syncs DNS records from provider
func SyncProvider(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid provider ID"})
		return
	}

	var provider models.Provider
	if err := database.DB.First(&provider, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Provider not found"})
		return
	}

	// Get provider service
	svc, err := getProviderService(&provider)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Sync records
	records, err := svc.SyncRecords()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to sync records: " + err.Error()})
		return
	}

	// Analyze records
	analysis := services.AnalyzeDNSRecords(records)

	// Log audit
	logAudit(c, models.ActionSync, models.ResourceTypeProvider, provider.ID, gin.H{
		"record_count":    len(records),
		"server_suggestions": len(analysis.ServerSuggestions),
	})

	c.JSON(http.StatusOK, gin.H{
		"records":            analysis.Records,
		"server_suggestions": analysis.ServerSuggestions,
	})
}

// testProviderConnection tests the provider connection
func testProviderConnection(provider *models.Provider) error {
	svc, err := getProviderService(provider)
	if err != nil {
		return err
	}

	return svc.TestConnection()
}

// getProviderService returns the appropriate provider service
func getProviderService(provider *models.Provider) (services.DNSProvider, error) {
	apiKey, err := crypto.Decrypt(provider.APIKey)
	if err != nil {
		return nil, err
	}

	var apiSecret string
	if provider.APISecret != "" {
		apiSecret, err = crypto.Decrypt(provider.APISecret)
		if err != nil {
			return nil, err
		}
	}

	switch provider.Name {
	case models.ProviderCloudflare:
		return services.NewCloudflareService(apiKey, apiSecret), nil
	case models.ProviderTencentCloud:
		return services.NewTencentCloudService(apiKey, apiSecret), nil
	default:
		return nil, nil
	}
}
