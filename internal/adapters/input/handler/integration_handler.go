package handler

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/myestatia/myestatia-go/internal/application/service"
)

type IntegrationHandler struct {
	service *service.IntegrationService
}

func NewIntegrationHandler(s *service.IntegrationService) *IntegrationHandler {
	return &IntegrationHandler{service: s}
}

type ConfigRequest struct {
	Settings  map[string]string `json:"settings"`
	CompanyID string            `json:"company_id"`
}

func (h *IntegrationHandler) TestConnection(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var req ConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err := h.service.TestConnection(r.Context(), id, req.Settings)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (h *IntegrationHandler) PreviewSync(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var req ConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	props, err := h.service.PreviewProperties(r.Context(), id, req.Settings)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(props)
}

// SyncProperties starts a full synchronization (asynchronous for large datasets)
func (h *IntegrationHandler) SyncProperties(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var req ConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if id != "resales" {
		http.Error(w, "Integration not supported", http.StatusBadRequest)
		return
	}

	if req.CompanyID == "" {
		http.Error(w, "company_id is required", http.StatusBadRequest)
		return
	}

	log.Printf("[INTEGRATION] Starting sync for %s, company %s", id, req.CompanyID)

	// Start sync asynchronously (for large datasets this can take 20-30 minutes)
	// CRITICAL: Use context.Background() instead of r.Context() because r.Context()
	// gets canceled when the HTTP response is sent, which would cancel the sync
	go func() {
		ctx := context.Background()
		result, err := h.service.SyncAllProperties(ctx, req.CompanyID, req.Settings)
		if err != nil {
			log.Printf("[INTEGRATION] Sync failed for company %s: %v", req.CompanyID, err)
			return
		}
		log.Printf("[INTEGRATION] Sync completed for company %s: %+v", req.CompanyID, result)
	}()

	// Return immediately with 202 Accepted
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "accepted",
		"message": "Synchronization started in the background",
	})
}

// GetSyncStatus returns the current sync status for a company
func (h *IntegrationHandler) GetSyncStatus(w http.ResponseWriter, r *http.Request) {
	companyID := r.URL.Query().Get("company_id")
	if companyID == "" {
		http.Error(w, "company_id query parameter is required", http.StatusBadRequest)
		return
	}

	status := h.service.GetSyncStatus(companyID)
	if status == nil {
		// No sync in progress or completed
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "idle",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}
