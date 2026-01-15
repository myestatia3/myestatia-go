package email

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

// ResendConfig holds Resend API configuration
type ResendConfig struct {
	APIKey string
	From   string
}

// LoadResendConfig loads Resend configuration from environment variables
func LoadResendConfig() ResendConfig {
	return ResendConfig{
		APIKey: os.Getenv("RESEND_API_KEY"),
		From:   os.Getenv("RESEND_FROM_EMAIL"),
	}
}

// IsValid checks if the Resend configuration has required fields
func (c ResendConfig) IsValid() bool {
	return c.APIKey != "" && c.From != ""
}

// ResendEmailSender handles sending emails via Resend API
type ResendEmailSender struct {
	config ResendConfig
}

// NewResendEmailSender creates a new Resend email sender
func NewResendEmailSender(config ResendConfig) *ResendEmailSender {
	return &ResendEmailSender{config: config}
}

type resendEmailRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	HTML    string   `json:"html"`
}

type resendErrorResponse struct {
	Message string `json:"message"`
	Name    string `json:"name"`
}

// SendEmail sends an email using Resend API
func (s *ResendEmailSender) SendEmail(to, subject, body string) error {
	if !s.config.IsValid() {
		return fmt.Errorf("invalid Resend configuration")
	}

	reqBody := resendEmailRequest{
		From:    s.config.From,
		To:      []string{to},
		Subject: subject,
		HTML:    body,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	log.Printf("[Resend] Sending email to %s from %s", to, s.config.From)

	req, err := http.NewRequest("POST", "https://api.resend.com/emails", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.config.APIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[Resend] ERROR: Request failed: %v", err)
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp resendErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil {
			log.Printf("[Resend] ERROR: API error (status %d): %s", resp.StatusCode, errResp.Message)
			return fmt.Errorf("resend API error: %s", errResp.Message)
		}
		log.Printf("[Resend] ERROR: API error (status %d)", resp.StatusCode)
		return fmt.Errorf("resend API error: status %d", resp.StatusCode)
	}

	log.Printf("[Resend] SUCCESS: Email sent to %s", to)
	return nil
}

// SendPasswordResetEmail sends a password reset email via Resend
func (s *ResendEmailSender) SendPasswordResetEmail(to, token string) error {
	// TODO: Get frontend URL from environment variable
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:5173"
	}

	resetURL := fmt.Sprintf("%s/reset-password/%s", frontendURL, token)

	subject := "Reset Your Password - MyEstatia"
	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); color: white; padding: 30px; text-align: center; border-radius: 10px 10px 0 0; }
        .content { background: #f9f9f9; padding: 30px; border-radius: 0 0 10px 10px; }
        .button { display: inline-block; padding: 12px 30px; background: #667eea; color: white; text-decoration: none; border-radius: 5px; margin: 20px 0; }
        .footer { text-align: center; margin-top: 20px; font-size: 12px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Reset Your Password</h1>
        </div>
        <div class="content">
            <p>Hello,</p>
            <p>We received a request to reset the password for your MyEstatia account.</p>
            <p>Click the button below to create a new password:</p>
            <p style="text-align: center;">
                <a href="%s" class="button">Reset Password</a>
            </p>
            <p>Or copy and paste this link into your browser:</p>
            <p style="word-break: break-all; color: #667eea;">%s</p>
            <p><strong>This link will expire in 2 hours.</strong></p>
            <p>If you didn't request a password reset, you can safely ignore this email.</p>
            <div class="footer">
                <p>© 2025 MyEstatia. All rights reserved.</p>
            </div>
        </div>
    </div>
</body>
</html>
`, resetURL, resetURL)

	return s.SendEmail(to, subject, body)
}
