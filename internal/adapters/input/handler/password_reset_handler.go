package handler

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/myestatia/myestatia-go/internal/application/service"
	"github.com/myestatia/myestatia-go/internal/domain/entity"
	"github.com/myestatia/myestatia-go/internal/infrastructure/repository"
	"golang.org/x/crypto/bcrypt"
)

// EmailSender interface for sending password reset emails
type EmailSender interface {
	SendPasswordResetEmail(to, token string) error
}

type PasswordResetHandler struct {
	agentService      *service.AgentService
	passwordResetRepo repository.PasswordResetRepository
	emailSender       EmailSender
}

func NewPasswordResetHandler(
	agentService *service.AgentService,
	passwordResetRepo repository.PasswordResetRepository,
	emailSender EmailSender,
) *PasswordResetHandler {
	return &PasswordResetHandler{
		agentService:      agentService,
		passwordResetRepo: passwordResetRepo,
		emailSender:       emailSender,
	}
}

type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

type ResetPasswordRequest struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}

type GenericResponse struct {
	Message string `json:"message"`
}

// ForgotPassword handles password reset requests
func (h *PasswordResetHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req ForgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Normalize email to lowercase
	req.Email = strings.ToLower(req.Email)

	// Always return the same message to prevent user enumeration
	genericMessage := "Si el usuario existe, recibirá un email para recuperar su contraseña. Recuerde revisar la carpeta de spam."

	// Find agent by email
	agent, err := h.agentService.GetByEmail(r.Context(), req.Email)
	if err != nil || agent == nil {
		// Don't reveal if user exists or not
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(GenericResponse{Message: genericMessage})
		return
	}

	// Generate random token (64 character hex string)
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}
	token := hex.EncodeToString(tokenBytes)

	// Create password reset record (token stored directly, not hashed)
	passwordReset := &entity.PasswordReset{
		UserID:    agent.ID,
		Token:     token,
		ExpiresAt: time.Now().Add(2 * time.Hour),
		Used:      false,
	}

	if err := h.passwordResetRepo.Create(passwordReset); err != nil {
		http.Error(w, "Failed to create password reset", http.StatusInternalServerError)
		return
	}

	// Send email with the token
	if err := h.emailSender.SendPasswordResetEmail(agent.Email, token); err != nil {
		// Log error but don't reveal it to user
		// In production, you might want to log this
		log.Printf("[PasswordReset] ERROR: Failed to send email to %s: %v", agent.Email, err)
		// For now, still return success to prevent user enumeration
	} else {
		log.Printf("[PasswordReset] SUCCESS: Email sent to %s", agent.Email)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(GenericResponse{Message: genericMessage})
}

// ValidateResetToken validates if a reset token is valid
func (h *PasswordResetHandler) ValidateResetToken(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")
	if token == "" {
		http.Error(w, "Token is required", http.StatusBadRequest)
		return
	}

	// Find all password reset records (we need to check hashed tokens)
	// This is a simplified approach - in production, you might want to optimize this
	reset, err := h.findResetByToken(r.Context(), token)
	if err != nil || reset == nil {
		http.Error(w, "Invalid or expired token", http.StatusBadRequest)
		return
	}

	if !reset.IsValid() {
		http.Error(w, "Invalid or expired token", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(GenericResponse{Message: "Token is valid"})
}

// ResetPassword resets the password using a valid token
func (h *PasswordResetHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate password strength
	if len(req.Password) < 8 {
		http.Error(w, "Password must be at least 8 characters long", http.StatusBadRequest)
		return
	}

	// Find password reset by token
	reset, err := h.findResetByToken(r.Context(), req.Token)
	if err != nil || reset == nil {
		http.Error(w, "Invalid or expired token", http.StatusBadRequest)
		return
	}

	if !reset.IsValid() {
		http.Error(w, "Invalid or expired token", http.StatusBadRequest)
		return
	}

	// Get agent
	agent, err := h.agentService.FindByID(r.Context(), reset.UserID)
	if err != nil || agent == nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	// Update agent password using UpdatePartial
	if err := h.agentService.UpdatePartial(r.Context(), agent.ID, map[string]interface{}{
		"password": string(hashedPassword),
	}); err != nil {
		http.Error(w, "Failed to update password", http.StatusInternalServerError)
		return
	}

	// Mark token as used
	if err := h.passwordResetRepo.MarkAsUsed(r.Context(), reset.ID); err != nil {
		// Log error but don't fail the request
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(GenericResponse{Message: "Password reset successfully"})
}

// Helper function to find reset by token
func (h *PasswordResetHandler) findResetByToken(ctx context.Context, token string) (*entity.PasswordReset, error) {
	return h.passwordResetRepo.FindByToken(ctx, token)
}
