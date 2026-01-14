package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/myestatia/myestatia-go/internal/adapters/input/middleware"
	"github.com/myestatia/myestatia-go/internal/application/service"
	"github.com/myestatia/myestatia-go/internal/domain/entity"
)

type LeadHandler struct {
	Service *service.LeadService
}

func NewLeadHandler(s *service.LeadService) *LeadHandler {
	return &LeadHandler{Service: s}
}

// POST /api/v1/leads
func (h *LeadHandler) CreateLead(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Name         string  `json:"name"`
		Email        string  `json:"email"`
		Phone        string  `json:"phone"`
		Language     string  `json:"language"`
		Source       string  `json:"source"`
		Budget       float64 `json:"budget"`
		Zone         string  `json:"zone"`
		PropertyType string  `json:"propertyType"`
		PropertyID   string  `json:"propertyId"`
		CompanyID    string  `json:"companyId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	var propertyID *string
	if req.PropertyID != "" {
		propertyID = &req.PropertyID
	}

	companyID, ok := r.Context().Value(middleware.CompanyIDKey).(string)
	if !ok || companyID == "" {
		http.Error(w, "Unauthorized: invalid company context", http.StatusUnauthorized)
		return
	}

	lead := &entity.Lead{
		Name:         req.Name,
		Email:        req.Email,
		Phone:        req.Phone,
		Language:     req.Language,
		Source:       req.Source,
		Budget:       req.Budget,
		Zone:         req.Zone,
		PropertyType: req.PropertyType,
		PropertyID:   propertyID,
		CompanyID:    companyID,
	}

	createdLead, created, err := h.Service.Create(context.Background(), lead)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if created {
		w.WriteHeader(http.StatusCreated) // 201 si es nuevo
	} else {
		w.WriteHeader(http.StatusOK) // 200 si ya existía
	}
	_ = json.NewEncoder(w).Encode(createdLead)
}

// GET /api/v1/leads
func (h *LeadHandler) GetAllLeads(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	leads, err := h.Service.FindAll(context.Background())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(leads)
}

// GET /api/v1/leads/{id}
func (h *LeadHandler) GetLeadByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/v1/leads/")
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	lead, err := h.Service.FindByID(context.Background(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if lead == nil {
		http.Error(w, "lead not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(lead)
}

// PUT /api/v1/leads/{id}
func (h *LeadHandler) UpdateLead(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/v1/leads/")
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	// Fetch existing lead first to avoid overwriting with zero values
	existingLead, err := h.Service.FindByID(context.Background(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if existingLead == nil {
		http.Error(w, "lead not found", http.StatusNotFound)
		return
	}

	var req struct {
		Name         *string  `json:"name"`
		Email        *string  `json:"email"`
		Phone        *string  `json:"phone"`
		Language     *string  `json:"language"`
		Budget       *float64 `json:"budget"`
		Zone         *string  `json:"zone"`
		PropertyType *string  `json:"propertyType"`
		Status       *string  `json:"status"`     // Added status
		PropertyID   *string  `json:"propertyId"` // Added propertyId
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	// Partial update logic: we only update fields that are present (non-nil) in the request.
	// This ensures that we don't overwrite existing data with empty values when the frontend
	// sends a partial object (e.g., only updating 'status').
	if req.Name != nil {
		existingLead.Name = *req.Name
	}
	if req.Email != nil {
		existingLead.Email = *req.Email
	}
	if req.Phone != nil {
		existingLead.Phone = *req.Phone
	}
	if req.Language != nil {
		existingLead.Language = *req.Language
	}
	if req.Budget != nil {
		existingLead.Budget = *req.Budget
	}
	if req.Zone != nil {
		existingLead.Zone = *req.Zone
	}
	if req.PropertyType != nil {
		existingLead.PropertyType = *req.PropertyType
	}
	if req.Status != nil {
		existingLead.Status = entity.LeadStatus(*req.Status)
	}
	if req.PropertyID != nil {
		if *req.PropertyID != "" {
			existingLead.PropertyID = req.PropertyID
		} else {
			// Handle empty string as unset if needed, or ignore.
			// Currently assuming empty string means "unset" or invalid,
			// but for safety we can explicitly set to nil if empty string is passed.
			existingLead.PropertyID = nil
		}
	}

	if err := h.Service.Update(context.Background(), existingLead); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// DELETE /api/v1/leads/{id}
func (h *LeadHandler) DeleteLead(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/v1/leads/")
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	if err := h.Service.Delete(context.Background(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GET /api/leads/bycompany/{companyId}:
func (h *LeadHandler) GetLeadByCompanyId(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/v1/leads/bycompany/")
	if id == "" {
		http.Error(w, "missing company id", http.StatusBadRequest)
		return
	}

	leads, err := h.Service.FindByCompanyId(context.Background(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if leads == nil {
		http.Error(w, "Leads not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(leads)
}

// GET /api/leads/byproperty/{propertyId}:
func (h *LeadHandler) GetLeadByPropertyId(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/v1/leads/byproperty/")
	if id == "" {
		http.Error(w, "missing property id", http.StatusBadRequest)
		return
	}

	leads, err := h.Service.FindByPropertyId(context.Background(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if leads == nil {
		http.Error(w, "Leads not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(leads)

}
