package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/myestatia/myestatia-go/internal/application/service"
	"github.com/myestatia/myestatia-go/internal/domain/entity"
	"github.com/myestatia/myestatia-go/internal/infrastructure/security"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	agentService      *service.AgentService
	companyService    *service.CompanyService
	invitationService *service.InvitationService
}

func NewAuthHandler(agentService *service.AgentService, companyService *service.CompanyService, invitationService *service.InvitationService) *AuthHandler {
	return &AuthHandler{
		agentService:      agentService,
		companyService:    companyService,
		invitationService: invitationService,
	}
}

type RegisterRequest struct {
	Name        string `json:"name"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	CompanyName string `json:"company_name"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string       `json:"token"`
	Agent entity.Agent `json:"agent"`
}

// Register is now disabled - users must register via invitation
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Registration is by invitation only. Please request an invitation from your company.", http.StatusForbidden)
}

// RegisterWithToken handles registration with an invitation token
func (h *AuthHandler) RegisterWithToken(w http.ResponseWriter, r *http.Request) {
	// Extract token from URL path
	// Expected format: /api/v1/auth/register/{token}
	token := strings.TrimPrefix(r.URL.Path, "/api/v1/auth/register/")
	if token == "" || token == r.URL.Path {
		http.Error(w, "missing invitation token", http.StatusBadRequest)
		return
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	req.Email = strings.ToLower(req.Email)

	invitation, err := h.invitationService.ValidateInvitation(r.Context(), token)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Invalid invitation token", http.StatusNotFound)
			return
		}
		if strings.Contains(err.Error(), "already used") {
			http.Error(w, "This invitation has already been used", http.StatusGone)
			return
		}
		if strings.Contains(err.Error(), "expired") {
			http.Error(w, "This invitation has expired", http.StatusGone)
			return
		}
		http.Error(w, "Invalid invitation: "+err.Error(), http.StatusBadRequest)
		return
	}

	if invitation.Email != req.Email {
		http.Error(w, "Email does not match invitation", http.StatusBadRequest)
		return
	}

	existingAgent, _ := h.agentService.GetByEmail(r.Context(), req.Email)
	if existingAgent != nil {
		http.Error(w, "User with this email already exists", http.StatusConflict)
		return
	}

	company, err := h.companyService.FindByID(r.Context(), invitation.CompanyID)
	if err != nil || company == nil {
		http.Error(w, "Company not found", http.StatusInternalServerError)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	agent := &entity.Agent{
		Name:      req.Name,
		Email:     req.Email,
		Password:  string(hashedPassword),
		Role:      "agent", // Default role for invited users
		CompanyID: company.ID,
	}

	if _, _, err := h.agentService.Create(r.Context(), agent); err != nil {
		http.Error(w, "Failed to create agent: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Delete invitation after successful registration
	if err := h.invitationService.DeleteInvitation(r.Context(), token); err != nil {
		// Log error but don't fail the registration
		// The user is already created at this point
	}

	// Generate Token
	token, err = security.GenerateToken(agent.ID, agent.CompanyID, agent.Role)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(AuthResponse{Token: token, Agent: *agent})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	req.Email = strings.ToLower(req.Email)

	agent, err := h.agentService.GetByEmail(r.Context(), req.Email)
	if err != nil || agent == nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(agent.Password), []byte(req.Password)); err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := security.GenerateToken(agent.ID, agent.CompanyID, agent.Role)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(AuthResponse{Token: token, Agent: *agent})
}
