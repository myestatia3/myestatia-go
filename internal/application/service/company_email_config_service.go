package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/myestatia/myestatia-go/internal/domain/entity"
	"github.com/myestatia/myestatia-go/internal/infrastructure/email"
	"github.com/myestatia/myestatia-go/internal/infrastructure/repository"
	"github.com/myestatia/myestatia-go/internal/infrastructure/security"
)

type CompanyEmailConfigService struct {
	repo          repository.CompanyEmailConfigRepository
	encryptionKey string
}

func NewCompanyEmailConfigService(repo repository.CompanyEmailConfigRepository, encryptionKey string) *CompanyEmailConfigService {
	return &CompanyEmailConfigService{
		repo:          repo,
		encryptionKey: encryptionKey,
	}
}

// GetEncryptionKey returns the encryption key (needed by worker manager)
func (s *CompanyEmailConfigService) GetEncryptionKey() string {
	return s.encryptionKey
}

func (s *CompanyEmailConfigService) CreateConfig(ctx context.Context, companyID, imapHost, imapUsername, imapPassword string, imapPort, pollIntervalSecs int, inboxFolder string) (*entity.CompanyEmailConfig, error) {
	// Check if config already exists
	existing, err := s.repo.FindByCompanyID(ctx, companyID)
	if err != nil {
		return nil, fmt.Errorf("error checking existing config: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("email configuration already exists for this company")
	}

	// Encrypt password
	encryptedPassword, err := security.Encrypt(imapPassword, s.encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt password: %w", err)
	}

	// Set defaults
	if inboxFolder == "" {
		inboxFolder = "INBOX"
	}
	if pollIntervalSecs <= 0 {
		pollIntervalSecs = 300 // 5 minutes default
	}
	if imapPort <= 0 {
		imapPort = 993
	}

	config := &entity.CompanyEmailConfig{
		CompanyID:        companyID,
		IMAPHost:         imapHost,
		IMAPPort:         imapPort,
		IMAPUsername:     imapUsername,
		IMAPPassword:     encryptedPassword,
		InboxFolder:      inboxFolder,
		PollIntervalSecs: pollIntervalSecs,
		IsEnabled:        true,
	}

	if err := s.repo.Create(config); err != nil {
		return nil, fmt.Errorf("failed to create config: %w", err)
	}

	log.Printf("[CompanyEmailConfigService] Created email config for company %s", companyID)
	return config, nil
}

func (s *CompanyEmailConfigService) UpdateConfig(ctx context.Context, id, imapHost, imapUsername, imapPassword string, imapPort, pollIntervalSecs int, inboxFolder string) (*entity.CompanyEmailConfig, error) {
	config, err := s.repo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("error finding config: %w", err)
	}
	if config == nil {
		return nil, fmt.Errorf("config not found")
	}

	// Update fields
	config.IMAPHost = imapHost
	config.IMAPPort = imapPort
	config.IMAPUsername = imapUsername
	config.InboxFolder = inboxFolder
	config.PollIntervalSecs = pollIntervalSecs

	// Only encrypt and update password if a new one is provided
	if imapPassword != "" {
		encryptedPassword, err := security.Encrypt(imapPassword, s.encryptionKey)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt password: %w", err)
		}
		config.IMAPPassword = encryptedPassword
	}

	if err := s.repo.Update(config); err != nil {
		return nil, fmt.Errorf("failed to update config: %w", err)
	}

	log.Printf("[CompanyEmailConfigService] Updated email config %s for company %s", id, config.CompanyID)
	return config, nil
}

func (s *CompanyEmailConfigService) GetConfigByCompanyID(ctx context.Context, companyID string) (*entity.CompanyEmailConfig, error) {
	return s.repo.FindByCompanyID(ctx, companyID)
}

func (s *CompanyEmailConfigService) GetAllEnabledConfigs(ctx context.Context) ([]*entity.CompanyEmailConfig, error) {
	return s.repo.FindAllEnabled(ctx)
}

func (s *CompanyEmailConfigService) DeleteConfig(ctx context.Context, id string) error {
	config, err := s.repo.FindByID(id)
	if err != nil {
		return fmt.Errorf("error finding config: %w", err)
	}
	if config == nil {
		return fmt.Errorf("config not found")
	}

	if err := s.repo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete config: %w", err)
	}

	log.Printf("[CompanyEmailConfigService] Deleted email config %s for company %s", id, config.CompanyID)
	return nil
}

func (s *CompanyEmailConfigService) ToggleEnabled(ctx context.Context, id string, enabled bool) error {
	config, err := s.repo.FindByID(id)
	if err != nil {
		return fmt.Errorf("error finding config: %w", err)
	}
	if config == nil {
		return fmt.Errorf("config not found")
	}

	config.IsEnabled = enabled
	if err := s.repo.Update(config); err != nil {
		return fmt.Errorf("failed to toggle config: %w", err)
	}

	status := "disabled"
	if enabled {
		status = "enabled"
	}
	log.Printf("[CompanyEmailConfigService] %s email config %s for company %s", status, id, config.CompanyID)
	return nil
}

func (s *CompanyEmailConfigService) TestConnection(ctx context.Context, config *entity.CompanyEmailConfig) error {
	// Decrypt password
	password, err := security.Decrypt(config.IMAPPassword, s.encryptionKey)
	if err != nil {
		return fmt.Errorf("failed to decrypt password: %w", err)
	}

	// Create temporary IMAP config
	imapConfig := email.Config{
		IMAPHost:    config.IMAPHost,
		IMAPPort:    config.IMAPPort,
		Username:    config.IMAPUsername,
		Password:    password,
		InboxFolder: config.InboxFolder,
	}

	// Try to connect
	client, err := email.NewIMAPClient(imapConfig)
	if err != nil {
		return fmt.Errorf("failed to create IMAP client: %w", err)
	}
	if err := client.Connect(); err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer client.Disconnect()

	log.Printf("[CompanyEmailConfigService] Connection test successful for company %s", config.CompanyID)
	return nil
}

func (s *CompanyEmailConfigService) UpdateLastSync(ctx context.Context, id string) error {
	return s.repo.UpdateLastSync(ctx, id, time.Now())
}

func (s *CompanyEmailConfigService) DecryptPassword(encryptedPassword string) (string, error) {
	return security.Decrypt(encryptedPassword, s.encryptionKey)
}

// CreateOAuth2Config creates a new email configuration using OAuth2
func (s *CompanyEmailConfigService) CreateOAuth2Config(
	ctx context.Context,
	companyID string,
	provider string, // "google"
	encryptedAccessToken string,
	encryptedRefreshToken string,
	tokenExpiry time.Time,
) (*entity.CompanyEmailConfig, error) {
	// Check if config already exists
	existing, err := s.repo.FindByCompanyID(ctx, companyID)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("email configuration already exists for this company")
	}

	config := &entity.CompanyEmailConfig{
		CompanyID:        companyID,
		AuthMethod:       "oauth2",
		OAuth2Provider:   provider,
		AccessToken:      encryptedAccessToken,
		RefreshToken:     encryptedRefreshToken,
		TokenExpiry:      &tokenExpiry,
		InboxFolder:      "INBOX",
		PollIntervalSecs: 300,
		IsEnabled:        true,
	}

	err = s.repo.Create(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create OAuth2 config: %w", err)
	}

	return config, nil
}

// UpdateToOAuth2 updates an existing configuration to use OAuth2
func (s *CompanyEmailConfigService) UpdateToOAuth2(
	ctx context.Context,
	id string,
	provider string,
	encryptedAccessToken string,
	encryptedRefreshToken string,
	tokenExpiry time.Time,
) error {
	config, err := s.repo.FindByID(id)
	if err != nil {
		return fmt.Errorf("configuration not found: %w", err)
	}

	// Update to OAuth2
	config.AuthMethod = "oauth2"
	config.OAuth2Provider = provider
	config.AccessToken = encryptedAccessToken
	config.RefreshToken = encryptedRefreshToken
	config.TokenExpiry = &tokenExpiry

	// Clear IMAP password (no longer needed)
	config.IMAPPassword = ""

	err = s.repo.Update(config)
	if err != nil {
		return fmt.Errorf("failed to update to OAuth2: %w", err)
	}

	return nil
}

// RefreshOAuth2Token refreshes an expired OAuth2 token
func (s *CompanyEmailConfigService) RefreshOAuth2Token(
	ctx context.Context,
	id string,
	newEncryptedAccessToken string,
	newEncryptedRefreshToken string,
	newTokenExpiry time.Time,
) error {
	config, err := s.repo.FindByID(id)
	if err != nil {
		return fmt.Errorf("configuration not found: %w", err)
	}

	if config.AuthMethod != "oauth2" {
		return fmt.Errorf("configuration is not using OAuth2")
	}

	config.AccessToken = newEncryptedAccessToken
	config.RefreshToken = newEncryptedRefreshToken
	config.TokenExpiry = &newTokenExpiry

	err = s.repo.Update(config)
	if err != nil {
		return fmt.Errorf("failed to refresh token: %w", err)
	}

	return nil
}

// DecryptOAuth2Tokens decrypts OAuth2 tokens for use by workers
func (s *CompanyEmailConfigService) DecryptOAuth2Tokens(config *entity.CompanyEmailConfig) (accessToken string, refreshToken string, err error) {
	if config.AuthMethod != "oauth2" {
		return "", "", fmt.Errorf("configuration is not using OAuth2")
	}

	accessToken, err = security.Decrypt(config.AccessToken, s.encryptionKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to decrypt access token: %w", err)
	}

	refreshToken, err = security.Decrypt(config.RefreshToken, s.encryptionKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to decrypt refresh token: %w", err)
	}

	return accessToken, refreshToken, nil
}
