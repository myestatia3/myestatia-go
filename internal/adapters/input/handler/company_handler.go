package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"bitbucket.org/statia/server/internal/application/service"
	"bitbucket.org/statia/server/internal/domain/entity"
)

type CompanyHandler struct {
	Service *service.CompanyService
}

func NewCompanyHandler(s *service.CompanyService) *CompanyHandler {
	return &CompanyHandler{Service: s}
}

// POST /api/v1/companies
func (h *CompanyHandler) CreateCompany(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req entity.Company
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.City == "" || req.Email1 == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}
	req.CreatedAt = time.Now()
	req.UpdatedAt = time.Now()

	createdCompany, created, err := h.Service.Create(context.Background(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if created {
		w.WriteHeader(http.StatusCreated) // 201 si es nueva
	} else {
		w.WriteHeader(http.StatusOK) // 200 si ya existía
	}
	_ = json.NewEncoder(w).Encode(createdCompany)
}

// GET /api/v1/companies
func (h *CompanyHandler) GetAllCompanies(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	companies, err := h.Service.FindAll(context.Background())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(companies)
}

// GET /api/v1/companies/{id}
func (h *CompanyHandler) GetCompanyByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/v1/companies/")
	if id == "" {
		http.Error(w, "Missing id", http.StatusBadRequest)
		return
	}

	company, err := h.Service.FindByID(context.Background(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if company == nil {
		http.Error(w, "Company not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(company)
}

// PUT /api/v1/companies/{id}
func (h *CompanyHandler) UpdateCompany(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/v1/companies/")
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	req["id"] = id
	req["updated_at"] = time.Now()

	if err := h.Service.UpdatePartial(context.Background(), id, req); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// DELETE /api/v1/companies/{id}
func (h *CompanyHandler) DeleteCompany(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/v1/companies/")
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
