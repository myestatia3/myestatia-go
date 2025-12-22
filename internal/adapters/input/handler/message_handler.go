package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/myestatia/myestatia-go/internal/application/service"
)

type MessageHandler struct {
	Service *service.MessageService
}

func NewMessageHandler(s *service.MessageService) *MessageHandler {
	return &MessageHandler{Service: s}
}

// Conversation DTOs
type MessageDTO struct {
	ID         string `json:"id"`
	SenderType string `json:"senderType"`
	Content    string `json:"content"`
	Timestamp  string `json:"timestamp"`
	Channel    string `json:"channel,omitempty"`
}

type ConversationDTO struct {
	ID        string       `json:"id"`
	LeadID    string       `json:"leadId"`
	Channel   string       `json:"channel,omitempty"`
	Messages  []MessageDTO `json:"messages"`
	CreatedAt string       `json:"createdAt"`
}

// GET /api/v1/lead/{id}/conversations
func (h *MessageHandler) GetConversations(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	messages, err := h.Service.GetMessagesByLeadID(context.Background(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Wrap in Conversation structure
	msgsDTO := make([]MessageDTO, len(messages))
	for i, m := range messages {
		msgsDTO[i] = MessageDTO{
			ID:         m.ID,
			SenderType: string(m.SenderType),
			Content:    m.Content,
			Timestamp:  m.Timestamp.Format(time.RFC3339),
		}
	}

	response := []ConversationDTO{}
	// If we have messages, we wrap them in a conversation object
	// Even if empty, we might return an empty list of conversations
	if len(messages) > 0 {
		response = append(response, ConversationDTO{
			ID:        "conv-" + id,
			LeadID:    id,
			Messages:  msgsDTO,
			CreatedAt: messages[0].Timestamp.Format(time.RFC3339),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// POST /api/v1/conversations/{leadId}/messages
func (h *MessageHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	leadId := r.PathValue("leadId")
	if leadId == "" {
		http.Error(w, "missing leadId", http.StatusBadRequest)
		return
	}

	var req struct {
		SenderType string `json:"senderType"`
		Content    string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	msg, err := h.Service.CreateMessage(context.Background(), leadId, req.SenderType, req.Content)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Map back to DTO
	res := MessageDTO{
		ID:         msg.ID,
		SenderType: string(msg.SenderType),
		Content:    msg.Content,
		Timestamp:  msg.Timestamp.Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}
