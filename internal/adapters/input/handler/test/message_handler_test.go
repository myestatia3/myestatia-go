package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/myestatia/myestatia-go/internal/adapters/input/handler"
	"github.com/myestatia/myestatia-go/internal/application/service"
	"github.com/myestatia/myestatia-go/internal/domain/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSendMessage_Handler(t *testing.T) {
	// GIVEN
	mockRepo := new(mocks.MessageRepositoryMock)
	svc := service.NewMessageService(mockRepo)
	h := handler.NewMessageHandler(svc)

	// Use a Mux to populate r.PathValue (Go 1.22+)
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/conversations/{leadId}/messages", h.SendMessage)

	msgReq := map[string]string{
		"senderType": "user",
		"content":    "Hello from Handler",
	}
	body, _ := json.Marshal(msgReq)

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/conversations/L1/messages", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	mockRepo.On("Create", mock.Anything).Return(nil)

	// WHEN - Call mux instead of handler directly
	mux.ServeHTTP(rr, req)

	// THEN
	assert.Equal(t, http.StatusOK, rr.Code)

	var response handler.MessageDTO
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Hello from Handler", response.Content)
}
