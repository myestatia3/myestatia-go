package handler

import (
	"encoding/json"
	"net/http"

	"github.com/myestatia/myestatia-go/internal/application/service"
	"github.com/myestatia/myestatia-go/internal/domain/entity"
	"github.com/myestatia/myestatia-go/internal/infrastructure/security"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	agentService   *service.AgentService
	companyService *service.CompanyService
}

func NewAuthHandler(agentService *service.AgentService, companyService *service.CompanyService) *AuthHandler {
	return &AuthHandler{
		agentService:   agentService,
		companyService: companyService,
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

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 1. Create Company
	company := &entity.Company{
		Name: req.CompanyName,
	}
	// Assuming CompanyService has Create method. Check next step.
	if _, _, err := h.companyService.Create(r.Context(), company); err != nil {
		http.Error(w, "Failed to create company: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 2. Hash Password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	// 3. Create Agent (linked to Company)
	agent := &entity.Agent{
		Name:      req.Name,
		Email:     req.Email,
		Password:  string(hashedPassword),
		Role:      "admin", // First user is admin
		CompanyID: company.ID,
	}

	if _, _, err := h.agentService.Create(r.Context(), agent); err != nil {
		http.Error(w, "Failed to create agent: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 4. Generate Token
	token, err := security.GenerateToken(agent.ID, agent.CompanyID, agent.Role)
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

	// 1. Find Agent by Email
	agent, err := h.agentService.GetByEmail(r.Context(), req.Email)
	if err != nil || agent == nil { // Check for nil if repo returns nil on not found
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// 2. Check Password
	if err := bcrypt.CompareHashAndPassword([]byte(agent.Password), []byte(req.Password)); err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// 3. Generate Token
	token, err := security.GenerateToken(agent.ID, agent.CompanyID, agent.Role)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(AuthResponse{Token: token, Agent: *agent})
}
