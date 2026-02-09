package handler

import (
	"encoding/json"
	"net/http"

	"github.com/myestatia/myestatia-go/internal/application/service"
)

type SchedulerHandler struct {
	schedulerService *service.SchedulerService
}

func NewSchedulerHandler(s *service.SchedulerService) *SchedulerHandler {
	return &SchedulerHandler{schedulerService: s}
}

type SyncConfigRequest struct {
	CompanyID string `json:"company_id"`
	APIKey    string `json:"api_key"`
	AgencyID  string `json:"agency_id"`
	Enabled   *bool  `json:"enabled,omitempty"`
}

// RegisterCompanySync registers a company for nightly sync
// POST /api/v1/scheduler/sync-config
func (h *SchedulerHandler) RegisterCompanySync(w http.ResponseWriter, r *http.Request) {
	var req SyncConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.CompanyID == "" || req.APIKey == "" || req.AgencyID == "" {
		http.Error(w, "company_id, api_key, and agency_id are required", http.StatusBadRequest)
		return
	}

	h.schedulerService.RegisterCompanySync(req.CompanyID, req.APIKey, req.AgencyID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "registered",
		"message": "Company registered for nightly sync at 3:00 AM",
	})
}

// UnregisterCompanySync removes a company from nightly sync
// DELETE /api/v1/scheduler/sync-config/{companyId}
func (h *SchedulerHandler) UnregisterCompanySync(w http.ResponseWriter, r *http.Request) {
	companyID := r.PathValue("companyId")
	if companyID == "" {
		http.Error(w, "companyId is required", http.StatusBadRequest)
		return
	}

	h.schedulerService.UnregisterCompanySync(companyID)

	w.WriteHeader(http.StatusNoContent)
}

// GetSyncConfig returns the sync configuration for a company
// GET /api/v1/scheduler/sync-config/{companyId}
func (h *SchedulerHandler) GetSyncConfig(w http.ResponseWriter, r *http.Request) {
	companyID := r.PathValue("companyId")
	if companyID == "" {
		http.Error(w, "companyId is required", http.StatusBadRequest)
		return
	}

	config := h.schedulerService.GetSyncConfig(companyID)
	if config == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"registered": false,
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"registered": true,
		"company_id": config.CompanyID,
		"enabled":    config.Enabled,
		// Not returning api_key for security
	})
}

// UpdateSyncConfig updates the sync configuration for a company
// PATCH /api/v1/scheduler/sync-config/{companyId}
func (h *SchedulerHandler) UpdateSyncConfig(w http.ResponseWriter, r *http.Request) {
	companyID := r.PathValue("companyId")
	if companyID == "" {
		http.Error(w, "companyId is required", http.StatusBadRequest)
		return
	}

	var req SyncConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Update enabled status if provided
	if req.Enabled != nil {
		h.schedulerService.SetSyncEnabled(companyID, *req.Enabled)
	}

	// If new credentials provided, update them
	if req.APIKey != "" && req.AgencyID != "" {
		h.schedulerService.RegisterCompanySync(companyID, req.APIKey, req.AgencyID)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "updated",
		"message": "Sync configuration updated",
	})
}

// GetAllSyncConfigs returns all registered sync configurations
// GET /api/v1/scheduler/sync-configs
func (h *SchedulerHandler) GetAllSyncConfigs(w http.ResponseWriter, r *http.Request) {
	configs := h.schedulerService.GetAllSyncConfigs()

	result := make([]map[string]interface{}, 0, len(configs))
	for _, config := range configs {
		result = append(result, map[string]interface{}{
			"company_id": config.CompanyID,
			"enabled":    config.Enabled,
			// Not returning api_key for security
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
