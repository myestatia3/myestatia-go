package oauth2

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
)

type OAuth2Config struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Config       *oauth2.Config
}

// NewOAuth2Config creates and returns the OAuth2 configuration for Google
func NewOAuth2Config(clientID, clientSecret, redirectURL string) *OAuth2Config {
	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes: []string{
			gmail.GmailReadonlyScope, // Read emails
			gmail.GmailModifyScope,   // Mark as read
		},
		Endpoint: google.Endpoint,
	}

	return &OAuth2Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Config:       config,
	}
}

// GetAuthURL returns the URL to redirect users to for OAuth2 authorization
func (c *OAuth2Config) GetAuthURL(state string) string {
	return c.Config.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
}

// ExchangeCode exchanges the authorization code for tokens
func (c *OAuth2Config) ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error) {
	token, err := c.Config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}
	return token, nil
}

// RefreshToken refreshes an expired access token using the refresh token
func (c *OAuth2Config) RefreshToken(ctx context.Context, refreshToken string) (*oauth2.Token, error) {
	token := &oauth2.Token{
		RefreshToken: refreshToken,
	}

	tokenSource := c.Config.TokenSource(ctx, token)
	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	return newToken, nil
}

// GetTokenFromRequest extracts the token from HTTP request (used for testing)
func (c *OAuth2Config) GetTokenFromRequest(r *http.Request) (*oauth2.Token, error) {
	var token oauth2.Token
	err := json.NewDecoder(r.Body).Decode(&token)
	if err != nil {
		return nil, fmt.Errorf("failed to decode token: %w", err)
	}
	return &token, nil
}

// ValidateToken checks if a token is still valid
func (c *OAuth2Config) ValidateToken(token *oauth2.Token) bool {
	if token == nil {
		return false
	}
	return token.Valid()
}

// LogTokenInfo logs token information (for debugging)
func (c *OAuth2Config) LogTokenInfo(token *oauth2.Token) {
	if token == nil {
		log.Println("[OAuth2] Token is nil")
		return
	}
	log.Printf("[OAuth2] Token valid: %v, Expiry: %v", token.Valid(), token.Expiry)
}
