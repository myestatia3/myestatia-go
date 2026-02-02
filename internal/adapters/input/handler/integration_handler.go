package handler

import (
	"encoding/json"
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
	Settings map[string]string `json:"settings"`
}

func (h *IntegrationHandler) TestConnection(w http.ResponseWriter, r *http.Request) {
	// path param: id
	// body: settings

	// Simple way to get ID from path would depend on the router (chi/mux/gin).
	// Assuming standard library or whatever is used.
	// Wait, the project uses net/http probably with some mux.
	// Let's check router.go later to see how params are extracted.
	// For now I'll assume I can parse URL path or pass it.

	// I will read path params in a generic way or rely on the caller to handle it if I used a framework.
	// Looking at other handlers...

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
