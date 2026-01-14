package worker

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/myestatia/myestatia-go/internal/application/service"
	"github.com/myestatia/myestatia-go/internal/domain/entity"
	"github.com/myestatia/myestatia-go/internal/infrastructure/email"
	repository "github.com/myestatia/myestatia-go/internal/infrastructure/repository"
)

// CompanyEmailWorker handles email polling for a single company
// Supports both IMAP (password auth) and Gmail API (OAuth2)
type CompanyEmailWorker struct {
	companyID        string
	companyName      string
	configID         string
	authMethod       string // "password" or "oauth2"
	emailLeadService *service.EmailLeadService

	// IMAP fields (for password auth)
	imapClient *email.IMAPClient
	imapConfig email.Config

	// Gmail API fields (for OAuth2)
	gmailClient *email.GmailClient

	processedEmailRepo repository.ProcessedEmailRepository // New field

	pollIntervalSecs int
}

// NewCompanyEmailWorker creates a new worker (detects auth method automatically)
func NewCompanyEmailWorker(
	companyID string,
	companyName string,
	configID string,
	authMethod string,
	emailLeadService *service.EmailLeadService,
	processedEmailRepo repository.ProcessedEmailRepository, // New arg
	imapConfig *email.Config, // nil for OAuth2
	gmailClient *email.GmailClient, // nil for IMAP
	pollIntervalSecs int,
) *CompanyEmailWorker {
	var safeIMAPConfig email.Config
	if imapConfig != nil {
		safeIMAPConfig = *imapConfig
	}

	return &CompanyEmailWorker{
		companyID:          companyID,
		companyName:        companyName,
		configID:           configID,
		authMethod:         authMethod,
		emailLeadService:   emailLeadService,
		processedEmailRepo: processedEmailRepo,
		imapConfig:         safeIMAPConfig,
		gmailClient:        gmailClient,
		pollIntervalSecs:   pollIntervalSecs,
	}
}

// Start begins the email polling loop
func (w *CompanyEmailWorker) Start(ctx context.Context) {
	log.Printf("[CompanyEmailWorker][%s] Starting worker for company: %s (auth: %s)",
		w.companyID, w.companyName, w.authMethod)

	// Initial poll
	w.pollEmails(ctx)

	// Create ticker for periodic polling
	ticker := time.NewTicker(time.Duration(w.pollIntervalSecs) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Printf("[CompanyEmailWorker][%s] Stopping worker", w.companyID)
			return
		case <-ticker.C:
			w.pollEmails(ctx)
		}
	}
}

// pollEmails fetches and processes emails based on auth method
func (w *CompanyEmailWorker) pollEmails(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[CompanyEmailWorker][%s] PANIC in pollEmails: %v", w.companyID, r)
		}
	}()

	log.Printf("[CompanyEmailWorker][%s] Polling inbox...", w.companyID)

	var emails []email.ParsedEmail
	var err error

	switch w.authMethod {
	case "oauth2":
		emails, err = w.fetchGmailEmails(ctx)
	case "password":
		emails, err = w.fetchIMAPEmails(ctx)
	default:
		log.Printf("[CompanyEmailWorker][%s] Unknown auth method: %s", w.companyID, w.authMethod)
		return
	}

	if err != nil {
		log.Printf("[CompanyEmailWorker][%s] Error fetching emails: %v", w.companyID, err)
		return
	}

	log.Printf("[CompanyEmailWorker][%s] Found %d unread emails", w.companyID, len(emails))

	// Process each email
	for _, email := range emails {
		// Check if already processed locally
		exists, err := w.processedEmailRepo.Exists(ctx, email.MessageID, w.companyID)
		if err != nil {
			log.Printf("[CompanyEmailWorker][%s] Error checking processed email %s: %v", w.companyID, email.MessageID, err)
			continue
		}
		if exists {
			continue
		}

		// Process the email (try to extract lead)
		processErr := w.emailLeadService.ProcessEmail(ctx, email)
		if processErr != nil {
			log.Printf("[CompanyEmailWorker][%s] Email processing result: %v", w.companyID, processErr)
			// We continue to save it as processed so we don't try again forever
			// specially for "unsupported source" or parser errors.
		}

		// Mark as processed locally ALWAYS
		processedEmail := &entity.ProcessedEmail{
			MessageID: email.MessageID,
			CompanyID: w.companyID,
		}
		if err := w.processedEmailRepo.Save(ctx, processedEmail); err != nil {
			log.Printf("[CompanyEmailWorker][%s] Error saving processed email record %s: %v", w.companyID, email.MessageID, err)
		}

		// DO NOT mark as read in Gmail/IMAP to respect user privacy

	}
}

// fetchGmailEmails fetches emails using Gmail API (OAuth2)
func (w *CompanyEmailWorker) fetchGmailEmails(ctx context.Context) ([]email.ParsedEmail, error) {
	if w.gmailClient == nil {
		return nil, fmt.Errorf("Gmail client not initialized")
	}

	gmailMessages, err := w.gmailClient.FetchUnreadEmails()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Gmail messages: %w", err)
	}

	// Convert Gmail messages to ParsedEmail format
	var emails []email.ParsedEmail
	for _, msg := range gmailMessages {
		emails = append(emails, email.ParsedEmail{
			MessageID: msg.MessageID,
			From:      msg.From,
			Subject:   msg.Subject,
			Body:      msg.Body,
			Date:      msg.Date,
		})
	}

	return emails, nil
}

// fetchIMAPEmails fetches emails using IMAP (password auth)
func (w *CompanyEmailWorker) fetchIMAPEmails(ctx context.Context) ([]email.ParsedEmail, error) {
	// Create fresh IMAP client for this poll
	imapClient, err := email.NewIMAPClient(w.imapConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create IMAP client: %w", err)
	}
	defer imapClient.Close()

	emails, err := imapClient.FetchUnreadEmails()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch IMAP emails: %w", err)
	}

	return emails, nil
}
