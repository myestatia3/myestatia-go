package handler

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/myestatia/myestatia-go/internal/application/service"
	googleoauth "github.com/myestatia/myestatia-go/internal/infrastructure/oauth2"
	"github.com/myestatia/myestatia-go/internal/infrastructure/security"
)

type GoogleOAuthHandler struct {
	oauth2Config       *googleoauth.OAuth2Config
	emailConfigService *service.CompanyEmailConfigService
	encryptionKey      string
}

func NewGoogleOAuthHandler(oauth2Config *googleoauth.OAuth2Config,
	emailConfigService *service.CompanyEmailConfigService, encryptionKey string) *GoogleOAuthHandler {
	return &GoogleOAuthHandler{
		oauth2Config:       oauth2Config,
		emailConfigService: emailConfigService,
		encryptionKey:      encryptionKey,
	}
}

// InitiateOAuth starts the OAuth2 flow
// GET /api/v1/auth/google/connect?company_id={id}
func (h *GoogleOAuthHandler) InitiateOAuth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	companyID := r.URL.Query().Get("company_id")
	if companyID == "" {
		http.Error(w, "company_id is required", http.StatusBadRequest)
		return
	}

	// Generate random state for CSRF protection
	state, err := generateState()
	if err != nil {
		log.Printf("[GoogleOAuth] Error generating state: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// Store state in session/cookie with company_id
	// For simplicity, we encode company_id in state (in production, use proper session)
	stateWithCompany := fmt.Sprintf("%s:%s", state, companyID)
	encodedState := base64.URLEncoding.EncodeToString([]byte(stateWithCompany))

	authURL := h.oauth2Config.GetAuthURL(encodedState)

	log.Printf("[GoogleOAuth] Redirecting to Google OAuth for company: %s", companyID)
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

// HandleCallback handles the OAuth2 callback from Google
// GET /api/v1/auth/google/callback?state=...&code=...
func (h *GoogleOAuthHandler) HandleCallback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract code and state
	code := r.URL.Query().Get("code")
	encodedState := r.URL.Query().Get("state")

	if code == "" || encodedState == "" {
		log.Println("[GoogleOAuth] Missing code or state in callback")
		http.Error(w, "missing code or state", http.StatusBadRequest)
		return
	}

	stateBytes, err := base64.URLEncoding.DecodeString(encodedState)
	if err != nil {
		log.Printf("[GoogleOAuth] Error decoding state: %v", err)
		http.Error(w, "invalid state", http.StatusBadRequest)
		return
	}

	stateStr := string(stateBytes)
	parts := strings.Split(stateStr, ":")
	if len(parts) != 2 {
		log.Printf("[GoogleOAuth] Invalid state format: %s", stateStr)
		http.Error(w, "invalid state", http.StatusBadRequest)
		return
	}
	companyID := parts[1]

	if companyID == "" {
		log.Println("[GoogleOAuth] Company ID not found in state")
		http.Error(w, "invalid state", http.StatusBadRequest)
		return
	}

	// Exchange code for tokens
	ctx := context.Background()
	token, err := h.oauth2Config.ExchangeCode(ctx, code)
	if err != nil {
		log.Printf("[GoogleOAuth] Error exchanging code: %v", err)
		http.Error(w, "failed to exchange code", http.StatusInternalServerError)
		return
	}

	log.Printf("[GoogleOAuth] Successfully obtained tokens for company: %s", companyID)
	h.oauth2Config.LogTokenInfo(token)

	// Encrypt tokens before storing
	encryptedAccessToken, err := security.Encrypt(token.AccessToken, h.encryptionKey)
	if err != nil {
		log.Printf("[GoogleOAuth] Error encrypting access token: %v", err)
		http.Error(w, "failed to encrypt tokens", http.StatusInternalServerError)
		return
	}

	encryptedRefreshToken, err := security.Encrypt(token.RefreshToken, h.encryptionKey)
	if err != nil {
		log.Printf("[GoogleOAuth] Error encrypting refresh token: %v", err)
		http.Error(w, "failed to encrypt tokens", http.StatusInternalServerError)
		return
	}

	// Check if config already exists
	existingConfig, err := h.emailConfigService.GetConfigByCompanyID(ctx, companyID)

	if existingConfig != nil {
		// Update existing config to use OAuth2
		log.Printf("[GoogleOAuth] Updating existing config to OAuth2 for company: %s", companyID)
		err = h.emailConfigService.UpdateToOAuth2(
			ctx,
			existingConfig.ID,
			"google",
			encryptedAccessToken,
			encryptedRefreshToken,
			token.Expiry,
		)
	} else {
		// Create new OAuth2 config
		log.Printf("[GoogleOAuth] Creating new OAuth2 config for company: %s", companyID)
		_, err = h.emailConfigService.CreateOAuth2Config(
			ctx,
			companyID,
			"google",
			encryptedAccessToken,
			encryptedRefreshToken,
			token.Expiry,
		)
	}

	if err != nil {
		log.Printf("[GoogleOAuth] Error saving OAuth2 config: %v", err)
		http.Error(w, "failed to save configuration", http.StatusInternalServerError)
		return
	}

	// Return success HTML that closes popup and notifies parent window
	successHTML := `
<!DOCTYPE html>
<html>
<head>
    <title>Authorization Successful</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            display: flex;
            justify-content: center;
            align-items: center;
            height: 100vh;
            margin: 0;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
        }
        .container {
            text-align: center;
        }
        h1 { font-size: 2em; margin-bottom: 20px; }
        p { font-size: 1.2em; }
    </style>
</head>
<body>
    <div class="container">
        <h1>✓ Authorization Successful!</h1>
        <p>Gmail connected successfully. This window will close automatically...</p>
    </div>
    <script>
        // Notify parent window
        if (window.opener) {
            window.opener.postMessage({ type: 'GOOGLE_AUTH_SUCCESS' }, '*');
        }
        // Close popup after 2 seconds
        setTimeout(() => window.close(), 2000);
    </script>
</body>
</html>
`

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(successHTML))
}

// DisconnectGmail disconnects Gmail by deleting OAuth2 tokens
// POST /api/v1/auth/google/disconnect
func (h *GoogleOAuthHandler) DisconnectGmail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		CompanyID string `json:"companyId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if req.CompanyID == "" {
		http.Error(w, "companyId is required", http.StatusBadRequest)
		return
	}

	ctx := context.Background()

	// Get existing config
	config, err := h.emailConfigService.GetConfigByCompanyID(ctx, req.CompanyID)
	if err != nil || config == nil {
		http.Error(w, "configuration not found", http.StatusNotFound)
		return
	}

	// Delete the configuration (user can recreate with IMAP if needed)
	err = h.emailConfigService.DeleteConfig(ctx, config.ID)
	if err != nil {
		log.Printf("[GoogleOAuth] Error disconnecting Gmail: %v", err)
		http.Error(w, "failed to disconnect", http.StatusInternalServerError)
		return
	}

	log.Printf("[GoogleOAuth] Successfully disconnected Gmail for company: %s", req.CompanyID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Gmail disconnected successfully",
	})
}

// generateState generates a random state string for CSRF protection
func generateState() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
