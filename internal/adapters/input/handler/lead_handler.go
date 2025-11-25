package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"bitbucket.org/statia/server/internal/application/service"
	"bitbucket.org/statia/server/internal/domain/entity"
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

	var req entity.Lead

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	createdLead, created, err := h.Service.Create(context.Background(), &req)
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

	var req struct {
		Name  string `json:"name"`
		Email string `json:"email"`
		Phone string `phone:"phone"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	lead := &entity.Lead{
		ID:    id,
		Name:  req.Name,
		Email: req.Email,
		Phone: req.Phone,
	}
	if err := h.Service.Update(context.Background(), lead); err != nil {
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
