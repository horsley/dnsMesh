package handlers

import (
	"dnsmesh/internal/database"
	"dnsmesh/internal/models"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetAuditLogs returns audit logs with pagination
func GetAuditLogs(c *gin.Context) {
	var logs []models.AuditLog

	// Build query
	query := database.DB.Order("created_at DESC")

	// Filter by resource type
	if resourceType := c.Query("resource_type"); resourceType != "" {
		query = query.Where("resource_type = ?", resourceType)
	}

	// Filter by action
	if action := c.Query("action"); action != "" {
		query = query.Where("action = ?", action)
	}

	// Pagination
	limit := 50
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 200 {
			limit = parsedLimit
		}
	}

	offset := 0
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	query = query.Limit(limit).Offset(offset)

	// Fetch logs
	if err := query.Find(&logs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch audit logs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"logs":  logs,
		"count": len(logs),
	})
}

// logAudit helper function to log audit events
func logAudit(c *gin.Context, action, resourceType string, resourceID uint, details gin.H) {
	detailsJSON, _ := json.Marshal(details)

	log := models.AuditLog{
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Details:      string(detailsJSON),
		IPAddress:    c.ClientIP(),
	}

	database.DB.Create(&log)
}
