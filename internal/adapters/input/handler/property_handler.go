package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/myestatia/myestatia-go/internal/adapters/input/middleware"
	"github.com/myestatia/myestatia-go/internal/application/service"
	"github.com/myestatia/myestatia-go/internal/domain/entity"
)

type PropertyHandler struct {
	Service        *service.PropertyService
	AgentService   *service.AgentService
	CompanyService *service.CompanyService
	Storage        service.StorageService
}

func NewPropertyHandler(s *service.PropertyService, as *service.AgentService, cs *service.CompanyService, storage service.StorageService) *PropertyHandler {
	return &PropertyHandler{
		Service:        s,
		AgentService:   as,
		CompanyService: cs,
		Storage:        storage,
	}
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

// GET /api/v1/public/properties/{id}
func (h *PropertyHandler) GetPublicPropertyByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/v1/public/properties/")
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

	contactPhone := ""
	// Try agent phone
	if property.CreatedByAgentID != nil && *property.CreatedByAgentID != "" {
		agent, err := h.AgentService.FindByID(r.Context(), *property.CreatedByAgentID)
		if err == nil && agent != nil && agent.Phone != "" {
			contactPhone = agent.Phone
		}
	}

	// Fallback to company phone
	if contactPhone == "" && property.CompanyID != "" {
		company, err := h.CompanyService.FindByID(r.Context(), property.CompanyID)
		if err == nil && company != nil {
			contactPhone = company.Phone1
		}
	}

	w.Header().Set("Content-Type", "application/json")
	resp := map[string]any{
		"property":     property,
		"contactPhone": contactPhone,
	}
	_ = json.NewEncoder(w).Encode(resp)
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

	// Limit upload size to 10MB
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		// Fallback to JSON if not multipart (for backward compatibility if needed, though client sends multipart now)
		// Try parsing as standard JSON body if multipart parsing fails/is empty and body exists
		if r.Body != http.NoBody {
			var req entity.Property
			if err := json.NewDecoder(r.Body).Decode(&req); err == nil {
				h.processCreateProperty(w, r, req)
				return
			}
		}
		http.Error(w, "failed to parse multipart form: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Extract JSON data from form
	jsonData := r.FormValue("data")
	var req entity.Property
	if jsonData != "" {
		if err := json.Unmarshal([]byte(jsonData), &req); err != nil {
			http.Error(w, "invalid json in 'data' field", http.StatusBadRequest)
			return
		}
	} else {
		// Try to decode entire body if no "data" field (though ParseMultipartForm consumes body)
		// If we are here, it's likely a multipart request. If 'data' is missing, it's an error.
		http.Error(w, "missing 'data' field", http.StatusBadRequest)
		return
	}

	// Handle Image Upload
	file, header, err := r.FormFile("image")
	if err == nil {
		defer file.Close()
		imageURL, err := h.Storage.UploadFile(r.Context(), file, header)
		if err != nil {
			http.Error(w, "failed to upload image: "+err.Error(), http.StatusInternalServerError)
			return
		}
		req.Image = imageURL
	}

	h.processCreateProperty(w, r, req)
}

func (h *PropertyHandler) processCreateProperty(w http.ResponseWriter, r *http.Request, req entity.Property) {

	companyID, ok := r.Context().Value(middleware.CompanyIDKey).(string)
	if !ok || companyID == "" {
		http.Error(w, "Unauthorized: invalid company context", http.StatusUnauthorized)
		return
	}
	req.CompanyID = companyID

	agentID, ok := r.Context().Value(middleware.AgentIDKey).(string)
	if ok && agentID != "" {
		req.CreatedByAgentID = &agentID
	}

	// Inject default values for now
	if req.Reference == "" {
		// Generate unique reference using timestamp and random string
		// Format: REF-YYYYMMDD-XXXX
		req.Reference = fmt.Sprintf("REF-%s-%s",
			time.Now().Format("20060102"),
			uuid.New().String()[:4])
	}
	req.Origin = entity.OriginManual

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
		SearchTerm: getQueryString(q, "q"),
		Status:     getQueryString(q, "status"),
		Origin:     getQueryString(q, "source"),
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

// GET /api/v1/property-subtypes
func (h *PropertyHandler) ListSubtypes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	propertyType := r.URL.Query().Get("type")

	subtypes, err := h.Service.ListSubtypes(r.Context(), propertyType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(subtypes)
}
