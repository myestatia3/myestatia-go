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

type AgentHandler struct {
	Service *service.AgentService
}

func NewAgentHandler(s *service.AgentService) *AgentHandler {
	return &AgentHandler{Service: s}
}

// POST /api/v1/agents
func (h *AgentHandler) CreateAgent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req entity.Agent
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.Email == "" || req.Phone == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}
	req.CreatedAt = time.Now()
	req.UpdatedAt = time.Now()

	createdAgent, created, err := h.Service.Create(context.Background(), &req)
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
	_ = json.NewEncoder(w).Encode(createdAgent)
}

// GET /api/v1/agents
func (h *AgentHandler) GetAllAgents(w http.ResponseWriter, r *http.Request) {
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

// GET /api/v1/agents/{id}
func (h *AgentHandler) GetAgentByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/v1/agents/")
	if id == "" {
		http.Error(w, "Missing id", http.StatusBadRequest)
		return
	}

	agent, err := h.Service.FindByID(context.Background(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if agent == nil {
		http.Error(w, "Agent not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(agent)
}

// PUT /api/v1/agents/{id}
func (h *AgentHandler) UpdateAgent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/v1/agents/")
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

// DELETE /api/v1/agents/{id}
func (h *AgentHandler) DeleteAgent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/v1/agents/")
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
