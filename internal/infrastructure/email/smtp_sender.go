package email

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"os"
)

// SMTPConfig holds SMTP email sending configuration
type SMTPConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
}

// LoadSMTPConfig loads SMTP configuration from environment variables
func LoadSMTPConfig() SMTPConfig {
	return SMTPConfig{
		Host:     os.Getenv("SMTP_HOST"),
		Port:     os.Getenv("SMTP_PORT"),
		Username: os.Getenv("SMTP_USERNAME"),
		Password: os.Getenv("SMTP_PASSWORD"),
		From:     os.Getenv("SMTP_FROM"),
	}
}

// IsValid checks if the SMTP configuration has required fields
func (c SMTPConfig) IsValid() bool {
	return c.Host != "" && c.Port != "" && c.Username != "" && c.Password != "" && c.From != ""
}

// EmailSender handles sending emails via SMTP
type EmailSender struct {
	config SMTPConfig
}

// NewEmailSender creates a new email sender
func NewEmailSender(config SMTPConfig) *EmailSender {
	return &EmailSender{config: config}
}

// SendEmail sends an email
func (s *EmailSender) SendEmail(to, subject, body string) error {
	if !s.config.IsValid() {
		return fmt.Errorf("invalid SMTP configuration")
	}

	// Prepare message
	message := []byte(fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"MIME-Version: 1.0\r\n"+
		"Content-Type: text/html; charset=UTF-8\r\n"+
		"\r\n"+
		"%s\r\n", s.config.From, to, subject, body))

	// Setup authentication
	auth := smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)

	// Connect to the server with TLS
	addr := fmt.Sprintf("%s:%s", s.config.Host, s.config.Port)

	// Try to send with TLS
	tlsConfig := &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         s.config.Host,
	}

	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		// Fallback to non-TLS
		return smtp.SendMail(addr, auth, s.config.From, []string{to}, message)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, s.config.Host)
	if err != nil {
		return err
	}
	defer client.Close()

	if err = client.Auth(auth); err != nil {
		return err
	}

	if err = client.Mail(s.config.From); err != nil {
		return err
	}

	if err = client.Rcpt(to); err != nil {
		return err
	}

	w, err := client.Data()
	if err != nil {
		return err
	}

	_, err = w.Write(message)
	if err != nil {
		return err
	}

	err = w.Close()
	if err != nil {
		return err
	}

	return client.Quit()
}

// SendPasswordResetEmail sends a password reset email
func (s *EmailSender) SendPasswordResetEmail(to, token string) error {
	// TODO: Get frontend URL from environment variable
	resetURL := fmt.Sprintf("http://localhost:5173/reset-password/%s", token)

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
