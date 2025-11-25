package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"bitbucket.org/statia/server/internal/application/service"
	"bitbucket.org/statia/server/internal/domain/entity"
)

type PropertyHandler struct {
	Service *service.PropertyService
}

func NewPropertyHandler(s *service.PropertyService) *PropertyHandler {
	return &PropertyHandler{Service: s}
}

// GET /api/v1/properties
func (h *PropertyHandler) GetAllProperties(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	properties, err := h.Service.GetAllProperties(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(properties)
}

// GET /api/v1/properties/{id}
func (h *PropertyHandler) GetPropertyByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/v1/properties/")
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	property, err := h.Service.GetPropertyByID(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if property == nil {
		http.Error(w, "Property not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(property)
}

// DELETE /api/v1/properties/{id}
func (h *PropertyHandler) DeleteProperty(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/v1/properties/")
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	if err := h.Service.DeleteProperty(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// POST /api/v1/properties
func (h *PropertyHandler) CreateProperty(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req entity.Property
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	createdProperty, created, err := h.Service.CreateProperty(r.Context(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if created {
		w.WriteHeader(http.StatusCreated)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	resp := map[string]any{
		"created":  created,
		"property": createdProperty,
	}
	_ = json.NewEncoder(w).Encode(resp)
}

// PUT /api/v1/properties/{id}
func (h *PropertyHandler) UpdateProperty(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/v1/properties/")
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	var req entity.Property
	req.ID = id

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if err := h.Service.UpdateProperty(context.Background(), &req); err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(req)
}

// GET /api/v1/properties/companies/{company_id}
func (h *PropertyHandler) GetPropertiesByCompany(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	companyID := strings.TrimPrefix(r.URL.Path, "/api/v1/properties/company/")

	if companyID == "" {
		http.Error(w, "missing company_id", http.StatusBadRequest)
		return
	}

	properties, err := h.Service.FindAllByCompanyID(context.Background(), companyID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(properties)
}

// GET /properties/search
func (h *PropertyHandler) SearchProperties(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	q := r.URL.Query()

	filter := entity.PropertyFilter{
		MinRooms:   getQueryInt(q, "minRooms"),
		MaxRooms:   getQueryInt(q, "maxRooms"),
		MinPrice:   getQueryFloat(q, "minBudget"),
		MaxPrice:   getQueryFloat(q, "maxBudget"),
		MinAreaM2:  getQueryInt(q, "minArea"),
		MaxAreaM2:  getQueryInt(q, "maxArea"),
		Province:   getQueryString(q, "province"),
		Address:    getQueryString(q, "address"),
		HasParking: getQueryBool(q, "parking"),
	}
	page := 1
	if p := getQueryInt(q, "page"); p != nil && *p > 0 {
		page = *p
	}

	limit := 10
	if l := getQueryInt(q, "limit"); l != nil && *l > 0 {
		limit = *l
	}

	filter.Limit = limit
	filter.Offset = (page - 1) * limit

	properties, err := h.Service.SearchProperties(r.Context(), filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(properties)
}
