package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/myestatia/myestatia-go/internal/application/service"
	"github.com/myestatia/myestatia-go/internal/domain/entity"
)

type CompanyEmailConfigHandler struct {
	emailConfigService *service.CompanyEmailConfigService
}

func NewCompanyEmailConfigHandler(emailConfigService *service.CompanyEmailConfigService) *CompanyEmailConfigHandler {
	return &CompanyEmailConfigHandler{
		emailConfigService: emailConfigService,
	}
}

type EmailConfigRequest struct {
	IMAPHost         string `json:"imapHost"`
	IMAPPort         int    `json:"imapPort"`
	IMAPUsername     string `json:"imapUsername"`
	IMAPPassword     string `json:"imapPassword"` // Only for create/update
	InboxFolder      string `json:"inboxFolder"`
	PollIntervalSecs int    `json:"pollIntervalSecs"`
}

type EmailConfigResponse struct {
	ID               string  `json:"id"`
	CompanyID        string  `json:"companyId"`
	AuthMethod       string  `json:"authMethod"` // "password" or "oauth2"
	IMAPHost         string  `json:"imapHost"`
	IMAPPort         int     `json:"imapPort"`
	IMAPUsername     string  `json:"imapUsername"`
	InboxFolder      string  `json:"inboxFolder"`
	PollIntervalSecs int     `json:"pollIntervalSecs"`
	IsEnabled        bool    `json:"isEnabled"`
	LastSyncAt       *string `json:"lastSyncAt"` // ISO format string
}

func toResponse(config *entity.CompanyEmailConfig) EmailConfigResponse {
	response := EmailConfigResponse{
		ID:               config.ID,
		CompanyID:        config.CompanyID,
		AuthMethod:       config.AuthMethod,
		IMAPHost:         config.IMAPHost,
		IMAPPort:         config.IMAPPort,
		IMAPUsername:     config.IMAPUsername,
		InboxFolder:      config.InboxFolder,
		PollIntervalSecs: config.PollIntervalSecs,
		IsEnabled:        config.IsEnabled,
	}

	if config.LastSyncAt != nil {
		syncTime := config.LastSyncAt.Format("2006-01-02T15:04:05Z07:00")
		response.LastSyncAt = &syncTime
	}

	return response
}

// CreateEmailConfig handles POST /api/companies/{id}/email-config
func (h *CompanyEmailConfigHandler) CreateEmailConfig(w http.ResponseWriter, r *http.Request) {
	// Extract company ID from URL path
	companyID := extractCompanyIDFromPath(r.URL.Path)
	if companyID == "" {
		http.Error(w, "Invalid company ID", http.StatusBadRequest)
		return
	}

	// TODO: Add authorization check - only agents from same company can create config
	// For now, we'll skip this until auth is implemented

	var req EmailConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.IMAPHost == "" || req.IMAPUsername == "" || req.IMAPPassword == "" {
		http.Error(w, "imapHost, imapUsername, and imapPassword are required", http.StatusBadRequest)
		return
	}

	config, err := h.emailConfigService.CreateConfig(
		r.Context(),
		companyID,
		req.IMAPHost,
		req.IMAPUsername,
		req.IMAPPassword,
		req.IMAPPort,
		req.PollIntervalSecs,
		req.InboxFolder,
	)
	if err != nil {
		log.Printf("[CompanyEmailConfigHandler] Error creating config: %v", err)
		if strings.Contains(err.Error(), "already exists") {
			http.Error(w, err.Error(), http.StatusConflict)
		} else {
			http.Error(w, "Failed to create email configuration", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(toResponse(config))
}

// GetEmailConfig handles GET /api/companies/{id}/email-config
func (h *CompanyEmailConfigHandler) GetEmailConfig(w http.ResponseWriter, r *http.Request) {
	companyID := extractCompanyIDFromPath(r.URL.Path)
	if companyID == "" {
		http.Error(w, "Invalid company ID", http.StatusBadRequest)
		return
	}

	// TODO: Add authorization check

	config, err := h.emailConfigService.GetConfigByCompanyID(r.Context(), companyID)
	if err != nil {
		log.Printf("[CompanyEmailConfigHandler] Error fetching config: %v", err)
		http.Error(w, "Failed to fetch email configuration", http.StatusInternalServerError)
		return
	}

	if config == nil {
		http.Error(w, "Email configuration not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(toResponse(config))
}

// UpdateEmailConfig handles PUT /api/companies/{id}/email-config
func (h *CompanyEmailConfigHandler) UpdateEmailConfig(w http.ResponseWriter, r *http.Request) {
	companyID := extractCompanyIDFromPath(r.URL.Path)
	if companyID == "" {
		http.Error(w, "Invalid company ID", http.StatusBadRequest)
		return
	}

	// TODO: Add authorization check

	// First, get existing config to get its ID
	existingConfig, err := h.emailConfigService.GetConfigByCompanyID(r.Context(), companyID)
	if err != nil {
		log.Printf("[CompanyEmailConfigHandler] Error fetching config: %v", err)
		http.Error(w, "Failed to fetch email configuration", http.StatusInternalServerError)
		return
	}
	if existingConfig == nil {
		http.Error(w, "Email configuration not found", http.StatusNotFound)
		return
	}

	var req EmailConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields (password is optional for update)
	if req.IMAPHost == "" || req.IMAPUsername == "" {
		http.Error(w, "imapHost and imapUsername are required", http.StatusBadRequest)
		return
	}

	config, err := h.emailConfigService.UpdateConfig(
		r.Context(),
		existingConfig.ID,
		req.IMAPHost,
		req.IMAPUsername,
		req.IMAPPassword, // Empty string means don't change password
		req.IMAPPort,
		req.PollIntervalSecs,
		req.InboxFolder,
	)
	if err != nil {
		log.Printf("[CompanyEmailConfigHandler] Error updating config: %v", err)
		http.Error(w, "Failed to update email configuration", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(toResponse(config))
}

// DeleteEmailConfig handles DELETE /api/companies/{id}/email-config
func (h *CompanyEmailConfigHandler) DeleteEmailConfig(w http.ResponseWriter, r *http.Request) {
	companyID := extractCompanyIDFromPath(r.URL.Path)
	if companyID == "" {
		http.Error(w, "Invalid company ID", http.StatusBadRequest)
		return
	}

	// TODO: Add authorization check

	// Get config to get its ID
	config, err := h.emailConfigService.GetConfigByCompanyID(r.Context(), companyID)
	if err != nil {
		log.Printf("[CompanyEmailConfigHandler] Error fetching config: %v", err)
		http.Error(w, "Failed to fetch email configuration", http.StatusInternalServerError)
		return
	}
	if config == nil {
		http.Error(w, "Email configuration not found", http.StatusNotFound)
		return
	}

	if err := h.emailConfigService.DeleteConfig(r.Context(), config.ID); err != nil {
		log.Printf("[CompanyEmailConfigHandler] Error deleting config: %v", err)
		http.Error(w, "Failed to delete email configuration", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// TestConnection handles POST /api/companies/{id}/email-config/test
func (h *CompanyEmailConfigHandler) TestConnection(w http.ResponseWriter, r *http.Request) {
	companyID := extractCompanyIDFromPath(r.URL.Path)
	if companyID == "" {
		http.Error(w, "Invalid company ID", http.StatusBadRequest)
		return
	}

	// TODO: Add authorization check

	config, err := h.emailConfigService.GetConfigByCompanyID(r.Context(), companyID)
	if err != nil {
		log.Printf("[CompanyEmailConfigHandler] Error fetching config: %v", err)
		http.Error(w, "Failed to fetch email configuration", http.StatusInternalServerError)
		return
	}
	if config == nil {
		http.Error(w, "Email configuration not found", http.StatusNotFound)
		return
	}

	if err := h.emailConfigService.TestConnection(r.Context(), config); err != nil {
		log.Printf("[CompanyEmailConfigHandler] Connection test failed: %v", err)
		response := map[string]interface{}{
			"success": false,
			"message": err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Connection successful",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ToggleEnabled handles PATCH /api/companies/{id}/email-config/toggle
func (h *CompanyEmailConfigHandler) ToggleEnabled(w http.ResponseWriter, r *http.Request) {
	companyID := extractCompanyIDFromPath(r.URL.Path)
	if companyID == "" {
		http.Error(w, "Invalid company ID", http.StatusBadRequest)
		return
	}

	// TODO: Add authorization check

	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	config, err := h.emailConfigService.GetConfigByCompanyID(r.Context(), companyID)
	if err != nil {
		log.Printf("[CompanyEmailConfigHandler] Error fetching config: %v", err)
		http.Error(w, "Failed to fetch email configuration", http.StatusInternalServerError)
		return
	}
	if config == nil {
		http.Error(w, "Email configuration not found", http.StatusNotFound)
		return
	}

	if err := h.emailConfigService.ToggleEnabled(r.Context(), config.ID, req.Enabled); err != nil {
		log.Printf("[CompanyEmailConfigHandler] Error toggling config: %v", err)
		http.Error(w, "Failed to toggle email configuration", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// extractCompanyIDFromPath extracts company ID from URL path
// Example: /api/companies/123-456/email-config -> "123-456"
func extractCompanyIDFromPath(path string) string {
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if part == "companies" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}
