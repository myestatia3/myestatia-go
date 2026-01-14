package email

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"time"

	"golang.org/x/oauth2"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"

	"github.com/myestatia/myestatia-go/internal/domain/entity"
	"github.com/myestatia/myestatia-go/internal/infrastructure/security"
)

// GmailClient handles Gmail API operations for OAuth2 authenticated accounts
type GmailClient struct {
	service       *gmail.Service
	config        *entity.CompanyEmailConfig
	encryptionKey string
	oauth2Config  *oauth2.Config
}

// NewGmailClient creates a new Gmail API client
func NewGmailClient(config *entity.CompanyEmailConfig, encryptionKey string, oauth2Cfg *oauth2.Config) (*GmailClient, error) {
	if config.AuthMethod != "oauth2" {
		return nil, fmt.Errorf("config is not using OAuth2 authentication")
	}

	// Decrypt tokens
	accessToken, err := security.Decrypt(config.AccessToken, encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt access token: %w", err)
	}

	refreshToken, err := security.Decrypt(config.RefreshToken, encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt refresh token: %w", err)
	}

	// Create OAuth2 token
	token := &oauth2.Token{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
	}
	if config.TokenExpiry != nil {
		token.Expiry = *config.TokenExpiry
	}

	// Create token source (handles auto-refresh)
	ctx := context.Background()
	tokenSource := oauth2Cfg.TokenSource(ctx, token)

	// Create Gmail service
	service, err := gmail.NewService(ctx, option.WithTokenSource(tokenSource))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gmail service: %w", err)
	}

	return &GmailClient{
		service:       service,
		config:        config,
		encryptionKey: encryptionKey,
		oauth2Config:  oauth2Cfg,
	}, nil
}

// Email represents a parsed Gmail message
type Email struct {
	MessageID string
	From      string
	Subject   string
	Body      string
	Date      time.Time
	IsUnread  bool
}

// FetchUnreadEmails fetches unread emails from the inbox
func (c *GmailClient) FetchUnreadEmails() ([]Email, error) {
	ctx := context.Background()

	// Query for unread messages in inbox
	query := "is:unread in:inbox"

	listCall := c.service.Users.Messages.List("me").Q(query).MaxResults(50)
	response, err := listCall.Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list messages: %w", err)
	}

	var emails []Email

	for _, msg := range response.Messages {
		// Get full message details
		fullMsg, err := c.service.Users.Messages.Get("me", msg.Id).Format("full").Do()
		if err != nil {
			log.Printf("[GmailClient] Error fetching message %s: %v", msg.Id, err)
			continue
		}

		email := c.parseMessage(fullMsg)
		emails = append(emails, email)
	}

	log.Printf("[GmailClient] Fetched %d unread emails", len(emails))
	return emails, nil
}

// parseMessage converts Gmail message to Email struct
func (c *GmailClient) parseMessage(msg *gmail.Message) Email {
	email := Email{
		MessageID: msg.Id,
		IsUnread:  true,
	}

	// Parse headers
	for _, header := range msg.Payload.Headers {
		switch header.Name {
		case "From":
			email.From = header.Value
		case "Subject":
			email.Subject = header.Value
		case "Date":
			// Parse date (simplified, could be improved)
			if parsedDate, err := time.Parse(time.RFC1123Z, header.Value); err == nil {
				email.Date = parsedDate
			}
		}
	}

	// Extract body
	email.Body = c.extractBody(msg.Payload)

	return email
}

// extractBody extracts the email body from the message payload
func (c *GmailClient) extractBody(payload *gmail.MessagePart) string {
	// If payload has body data, decode it
	if payload.Body != nil && payload.Body.Data != "" {
		decoded, err := base64.URLEncoding.DecodeString(payload.Body.Data)
		if err == nil {
			return string(decoded)
		}
	}

	// Check parts recursively
	if len(payload.Parts) > 0 {
		for _, part := range payload.Parts {
			// Prefer text/html, fallback to text/plain
			if part.MimeType == "text/html" || part.MimeType == "text/plain" {
				if part.Body != nil && part.Body.Data != "" {
					decoded, err := base64.URLEncoding.DecodeString(part.Body.Data)
					if err == nil {
						return string(decoded)
					}
				}
			}

			// Recursive check for multipart
			if len(part.Parts) > 0 {
				body := c.extractBody(part)
				if body != "" {
					return body
				}
			}
		}
	}

	return ""
}

// MarkAsRead marks an email as read
func (c *GmailClient) MarkAsRead(messageID string) error {
	ctx := context.Background()

	modifyRequest := &gmail.ModifyMessageRequest{
		RemoveLabelIds: []string{"UNREAD"},
	}

	_, err := c.service.Users.Messages.Modify("me", messageID, modifyRequest).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to mark message as read: %w", err)
	}

	return nil
}

// RefreshTokenIfNeeded checks if token needs refresh and updates the database
func (c *GmailClient) RefreshTokenIfNeeded(updateCallback func(newAccessToken, newRefreshToken string, newExpiry time.Time) error) error {
	// Token refresh is handled automatically by oauth2.TokenSource
	// We just need to get the refreshed token and save it
	ctx := context.Background()

	token := &oauth2.Token{
		RefreshToken: c.config.RefreshToken, // This is still encrypted
	}

	// Decrypt refresh token
	refreshToken, err := security.Decrypt(c.config.RefreshToken, c.encryptionKey)
	if err != nil {
		return fmt.Errorf("failed to decrypt refresh token: %w", err)
	}

	token.RefreshToken = refreshToken

	// Get fresh token (will refresh if expired)
	tokenSource := c.oauth2Config.TokenSource(ctx, token)
	newToken, err := tokenSource.Token()
	if err != nil {
		return fmt.Errorf("failed to refresh token: %w", err)
	}

	// If token was refreshed, encrypt and update
	if newToken.AccessToken != token.AccessToken {
		encryptedAccess, err := security.Encrypt(newToken.AccessToken, c.encryptionKey)
		if err != nil {
			return fmt.Errorf("failed to encrypt new access token: %w", err)
		}

		encryptedRefresh := c.config.RefreshToken // Refresh token typically doesn't change
		if newToken.RefreshToken != "" && newToken.RefreshToken != refreshToken {
			encryptedRefresh, err = security.Encrypt(newToken.RefreshToken, c.encryptionKey)
			if err != nil {
				return fmt.Errorf("failed to encrypt new refresh token: %w", err)
			}
		}

		// Call update callback
		if err := updateCallback(encryptedAccess, encryptedRefresh, newToken.Expiry); err != nil {
			return fmt.Errorf("failed to update tokens in database: %w", err)
		}

		log.Println("[GmailClient] Token refreshed successfully")
	}

	return nil
}
