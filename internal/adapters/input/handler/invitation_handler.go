package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/myestatia/myestatia-go/internal/application/service"
)

type InvitationHandler struct {
	Service *service.InvitationService
}

func NewInvitationHandler(s *service.InvitationService) *InvitationHandler {
	return &InvitationHandler{
		Service: s,
	}
}

// RequestInvitationRequest represents the request body for requesting an invitation
type RequestInvitationRequest struct {
	Email       string `json:"email"`
	CompanyName string `json:"companyName"`
}

// RequestInvitationResponse represents the response for requesting an invitation
type RequestInvitationResponse struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

// POST /api/v1/invitations/request
func (h *InvitationHandler) RequestInvitation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RequestInvitationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.CompanyName == "" {
		http.Error(w, "email and companyName are required", http.StatusBadRequest)
		return
	}

	// Request invitation
	err := h.Service.RequestInvitation(context.Background(), req.Email, req.CompanyName)
	if err != nil {
		if strings.Contains(err.Error(), "company not found") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(RequestInvitationResponse{
				Message: "If there are available invitations for " + req.CompanyName + ", a registration link will be sent to " + req.Email,
				Success: false,
			})
			return
		}

		if strings.Contains(err.Error(), "no invitations available") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(RequestInvitationResponse{
				Message: "No invitations available for " + req.CompanyName + ". Please contact myestatia@gmail.com to request more.",
				Success: false,
			})
			return
		}

		http.Error(w, "failed to request invitation: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(RequestInvitationResponse{
		Message: "If there are available invitations for " + req.CompanyName + ", a registration link will be sent to " + req.Email,
		Success: true,
	})
}

// GET /api/v1/invitations/validate/:token
func (h *InvitationHandler) ValidateInvitation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract token from URL
	token := strings.TrimPrefix(r.URL.Path, "/api/v1/invitations/validate/")
	if token == "" {
		http.Error(w, "missing token", http.StatusBadRequest)
		return
	}

	// Validate invitation
	invitation, err := h.Service.ValidateInvitation(context.Background(), token)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "invitation not found", http.StatusNotFound)
			return
		}
		if strings.Contains(err.Error(), "already used") {
			http.Error(w, "invitation already used", http.StatusGone)
			return
		}
		if strings.Contains(err.Error(), "expired") {
			http.Error(w, "invitation expired", http.StatusGone)
			return
		}
		http.Error(w, "invalid invitation: "+err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(invitation)
}
