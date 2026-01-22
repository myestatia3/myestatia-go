package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/myestatia/myestatia-go/internal/domain/entity"
	"github.com/myestatia/myestatia-go/internal/infrastructure/repository"
)

type EmailSender interface {
	SendInvitationEmail(to, token, companyName string) error
}

type InvitationService struct {
	invitationRepo repository.InvitationRepository
	companyRepo    repository.CompanyRepository
	emailSender    EmailSender
}

func NewInvitationService(
	invitationRepo repository.InvitationRepository,
	companyRepo repository.CompanyRepository,
	emailSender EmailSender,
) *InvitationService {
	return &InvitationService{
		invitationRepo: invitationRepo,
		companyRepo:    companyRepo,
		emailSender:    emailSender,
	}
}

func (s *InvitationService) RequestInvitation(ctx context.Context, email, companyName string) error {
	email = strings.ToLower(strings.TrimSpace(email))
	companyName = strings.TrimSpace(companyName)

	if email == "" || companyName == "" {
		return errors.New("email and company name are required")
	}

	company, err := s.companyRepo.FindByName(ctx, companyName)
	if err != nil {
		return fmt.Errorf("failed to find company: %w", err)
	}

	if company == nil {
		return errors.New("company not found")
	}

	if company.AvailableInvitations() <= 0 {
		return errors.New("no invitations available")
	}
	token, err := s.GenerateInvitationToken()
	if err != nil {
		return fmt.Errorf("failed to generate token: %w", err)
	}

	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	invitation := &entity.Invitation{
		Token:     token,
		Email:     email,
		CompanyID: company.ID,
		ExpiresAt: expiresAt,
	}

	if err := s.invitationRepo.Create(ctx, invitation); err != nil {
		return fmt.Errorf("failed to create invitation: %w", err)
	}

	company.UsedInvitations++
	if err := s.companyRepo.UpdatePartial(ctx, company.ID, map[string]interface{}{
		"used_invitations": company.UsedInvitations,
	}); err != nil {
		return fmt.Errorf("failed to update company: %w", err)
	}

	if err := s.emailSender.SendInvitationEmail(email, token, company.Name); err != nil {
		return fmt.Errorf("failed to send invitation email: %w", err)
	}

	return nil
}

func (s *InvitationService) ValidateInvitation(ctx context.Context, token string) (*entity.Invitation, error) {
	invitation, err := s.invitationRepo.FindByToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to find invitation: %w", err)
	}

	if invitation == nil {
		return nil, errors.New("invitation not found")
	}

	if !invitation.IsValid() {
		if invitation.Used {
			return nil, errors.New("invitation already used")
		}
		if invitation.IsExpired() {
			return nil, errors.New("invitation expired")
		}
	}

	return invitation, nil
}

// DeleteInvitation deletes an invitation after successful registration
func (s *InvitationService) DeleteInvitation(ctx context.Context, token string) error {
	return s.invitationRepo.Delete(ctx, token)
}
func (s *InvitationService) GenerateInvitationToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func (s *InvitationService) GetFrontendURL() string {
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:5173"
	}
	return frontendURL
}
