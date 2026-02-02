package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/myestatia/myestatia-go/internal/adapters/input/middleware"
	"github.com/myestatia/myestatia-go/internal/application/service"
)

type PresentationHandler struct {
	Service *service.PresentationService
}

func NewPresentationHandler(s *service.PresentationService) *PresentationHandler {
	return &PresentationHandler{
		Service: s,
	}
}

type CreatePresentationRequest struct {
	LeadId      string   `json:"leadId"`
	PropertyIds []string `json:"propertyIds"`
}
type CreatePresentationResponse struct {
	Token string `json:"token"`
	URL   string `json:"url"`
}

// POST /api/v1/presentations
func (h *PresentationHandler) CreatePresentation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreatePresentationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.LeadId == "" || len(req.PropertyIds) == 0 {
		http.Error(w, "leadId and propertyIds are required", http.StatusBadRequest)
		return
	}

	token, err := h.Service.GenerateToken(req.LeadId, req.PropertyIds)
	if err != nil {
		http.Error(w, "failed to generate presentation: "+err.Error(), http.StatusInternalServerError)
		return
	}

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" { //Esto podremos eliminarlo cuando tengamos todo bien asentado
		frontendURL = "http://localhost:5173"
	}

	presentationURL := frontendURL + "/presentations/" + token

	resp := CreatePresentationResponse{
		Token: token,
		URL:   presentationURL,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
}

// GET /api/v1/public/presentations/:token
func (h *PresentationHandler) GetPresentation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	token := strings.TrimPrefix(r.URL.Path, "/api/v1/public/presentations/")
	if token == "" {
		http.Error(w, "missing token", http.StatusBadRequest)
		return
	}

	presentation, err := h.Service.GetPresentation(r.Context(), token)
	if err != nil {
		if strings.Contains(err.Error(), "expired") {
			http.Error(w, "presentation has expired", http.StatusGone)
			return
		}
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "presentation not found", http.StatusNotFound)
			return
		}
		http.Error(w, "invalid presentation: "+err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(presentation)
}

// GET /api/v1/presentations/matching-properties/:leadId
func (h *PresentationHandler) GetMatchingProperties(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	leadID := strings.TrimPrefix(r.URL.Path, "/api/v1/presentations/matching-properties/")

	if leadID == "" {
		http.Error(w, "missing lead ID", http.StatusBadRequest)
		return
	}

	companyID, ok := r.Context().Value(middleware.CompanyIDKey).(string)
	if !ok || companyID == "" {
		http.Error(w, "unauthorized: invalid company context", http.StatusUnauthorized)
		return
	}

	matches, err := h.Service.GetMatchingProperties(context.Background(), leadID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "lead not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to get matching properties: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(matches)
}
