package worker

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/myestatia/myestatia-go/internal/application/service"
	"github.com/myestatia/myestatia-go/internal/domain/entity"
	"github.com/myestatia/myestatia-go/internal/infrastructure/email"
	repository "github.com/myestatia/myestatia-go/internal/infrastructure/repository"
	"golang.org/x/oauth2"
)

// EmailWorkerManager manages multiple company email workers
type EmailWorkerManager struct {
	emailConfigService *service.CompanyEmailConfigService
	propertyRepo       repository.PropertyRepository
	leadRepo           repository.LeadRepository
	processedEmailRepo repository.ProcessedEmailRepository // New repo
	workers            map[string]*CompanyEmailWorker      // key: companyID
	workerContexts     map[string]context.CancelFunc       // key: companyID
	wg                 sync.WaitGroup                      // Wait for all workers to finish
	mu                 sync.RWMutex
	configReloadSecs   int
}

// NewEmailWorkerManager creates a new worker manager
func NewEmailWorkerManager(
	emailConfigService *service.CompanyEmailConfigService,
	propertyRepo repository.PropertyRepository,
	leadRepo repository.LeadRepository,
	processedEmailRepo repository.ProcessedEmailRepository, // Add this
) *EmailWorkerManager {
	// Default poll interval for prod, can be overridden elsewhere
	// In dev we might want faster reload

	manager := &EmailWorkerManager{
		emailConfigService: emailConfigService,
		propertyRepo:       propertyRepo,
		leadRepo:           leadRepo,
		processedEmailRepo: processedEmailRepo,
		workers:            make(map[string]*CompanyEmailWorker),
		workerContexts:     make(map[string]context.CancelFunc),
		configReloadSecs:   600, // Reload configs every 10 minutes
	}

	// Check for dev environment to speed up config reload
	if os.Getenv("APP_ENV") == "development" {
		manager.configReloadSecs = 60
	}

	return manager
}

// Start begins the worker manager
func (m *EmailWorkerManager) Start(ctx context.Context) {
	log.Println("[EmailWorkerManager] Starting email worker manager")

	// Initial load
	m.reloadWorkers(ctx)

	// Periodically reload configurations
	ticker := time.NewTicker(time.Duration(m.configReloadSecs) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("[EmailWorkerManager] Shutting down email worker manager")
			m.stopAllWorkers()
			return
		case <-ticker.C:
			m.reloadWorkers(ctx)
		}
	}
}

// reloadWorkers loads all enabled email configurations and starts/stops workers
func (m *EmailWorkerManager) reloadWorkers(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[EmailWorkerManager] PANIC in reloadWorkers: %v", r)
		}
	}()

	log.Println("[EmailWorkerManager] Reloading email configurations...")

	// Get all enabled configs
	configs, err := m.emailConfigService.GetAllEnabledConfigs(ctx)
	if err != nil {
		log.Printf("[EmailWorkerManager] Error loading email configs: %v", err)
		return
	}

	log.Printf("[EmailWorkerManager] Found %d enabled email configuration(s)", len(configs))

	// Track active company IDs
	activeCompanyIDs := make(map[string]bool)

	// Start or update workers for each config
	for i, config := range configs {
		if config == nil {
			log.Printf("[EmailWorkerManager] Skipping nil config at index %d", i)
			continue
		}

		log.Printf("[EmailWorkerManager] Processing config for company %s (Auth: %s)", config.CompanyID, config.AuthMethod)
		activeCompanyIDs[config.CompanyID] = true

		m.mu.RLock()
		_, exists := m.workers[config.CompanyID]
		m.mu.RUnlock()

		if !exists {
			// Start new worker
			if err := m.startWorker(ctx, config); err != nil {
				log.Printf("[EmailWorkerManager] Failed to start worker for company %s: %v",
					config.CompanyID, err)
			}
		}
		// Note: If worker already exists, it continues running with its config
		// To update config, user would disable/enable or restart the application
	}

	// Stop workers for companies no longer enabled
	m.mu.RLock()
	currentWorkers := make(map[string]bool)
	for companyID := range m.workers {
		currentWorkers[companyID] = true
	}
	m.mu.RUnlock()

	for companyID := range currentWorkers {
		if !activeCompanyIDs[companyID] {
			m.stopWorker(companyID)
		}
	}
}

// startWorker starts a new worker for a company (supports both IMAP and Gmail API)
func (m *EmailWorkerManager) startWorker(parentCtx context.Context, config *entity.CompanyEmailConfig) error {
	companyName := "Unknown"
	if config.Company != nil {
		companyName = config.Company.Name
	}

	var worker *CompanyEmailWorker
	var err error

	// Create email lead service for this company (same for both auth methods)
	emailLeadService := service.NewEmailLeadService(
		m.propertyRepo,
		m.leadRepo,
		email.Config{DefaultCompanyID: config.CompanyID},
	)

	// Validate dependencies
	if m.emailConfigService == nil {
		return fmt.Errorf("emailConfigService is nil")
	}

	switch config.AuthMethod {
	case "oauth2":
		// Gmail API (OAuth2)
		log.Printf("[EmailWorkerManager] Creating OAuth2 worker for company %s", config.CompanyID)

		// Create OAuth2 config (reuse from main)
		clientID := os.Getenv("GOOGLE_CLIENT_ID")
		clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
		redirectURL := os.Getenv("GOOGLE_REDIRECT_URL")

		if clientID == "" || clientSecret == "" {
			log.Printf("[EmailWorkerManager] WARNING: Google OAuth credentials missing in environment")
		}

		oauth2Cfg := &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Scopes:       []string{"https://www.googleapis.com/auth/gmail.readonly", "https://www.googleapis.com/auth/gmail.modify"},
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://accounts.google.com/o/oauth2/auth",
				TokenURL: "https://oauth2.googleapis.com/token",
			},
		}

		// Create Gmail client
		encryptionKey := m.emailConfigService.GetEncryptionKey()
		if encryptionKey == "" {
			return fmt.Errorf("encryption key is empty")
		}

		gmailClient, err := email.NewGmailClient(config, encryptionKey, oauth2Cfg)
		if err != nil {
			return fmt.Errorf("failed to create Gmail client: %w", err)
		}

		worker = NewCompanyEmailWorker(
			config.CompanyID,
			companyName,
			config.ID,
			"oauth2",
			emailLeadService,
			m.processedEmailRepo, // New arg
			nil,                  // No IMAP config
			gmailClient,
			config.PollIntervalSecs,
		)

	case "password":
		// IMAP (password auth)
		log.Printf("[EmailWorkerManager] Creating IMAP worker for company %s", config.CompanyID)

		if config.IMAPPassword == "" {
			return fmt.Errorf("IMAP password is required for password auth method")
		}

		// Decrypt password
		password, err := m.emailConfigService.DecryptPassword(config.IMAPPassword)
		if err != nil {
			return fmt.Errorf("failed to decrypt password: %w", err)
		}

		// Create IMAP config
		imapConfig := &email.Config{
			IMAPHost:         config.IMAPHost,
			IMAPPort:         config.IMAPPort,
			Username:         config.IMAPUsername,
			Password:         password,
			InboxFolder:      config.InboxFolder,
			PollIntervalSecs: config.PollIntervalSecs,
			DefaultCompanyID: config.CompanyID,
		}

		worker = NewCompanyEmailWorker(
			config.CompanyID,
			companyName,
			config.ID,
			"password",
			emailLeadService,
			m.processedEmailRepo, // New arg
			imapConfig,
			nil, // No Gmail client
			config.PollIntervalSecs,
		)

	default:
		return fmt.Errorf("unsupported auth method: %s", config.AuthMethod)
	}

	if err != nil {
		return err
	}

	// Create context for this worker
	workerCtx, cancel := context.WithCancel(parentCtx)

	// Increment wait group
	m.wg.Add(1)

	// Start worker in goroutine
	go func() {
		defer m.wg.Done()
		worker.Start(workerCtx)
	}()

	// Store worker and cancel function
	m.mu.Lock()
	m.workers[config.CompanyID] = worker
	m.workerContexts[config.CompanyID] = cancel
	m.mu.Unlock()

	log.Printf("[EmailWorkerManager] Started %s worker for company %s (%s)", config.AuthMethod, config.CompanyID, companyName)
	return nil
}

// stopWorker stops a worker for a company
func (m *EmailWorkerManager) stopWorker(companyID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	cancel, exists := m.workerContexts[companyID]
	if !exists {
		return
	}

	// Cancel context to stop worker
	cancel()

	// Remove from maps
	delete(m.workers, companyID)
	delete(m.workerContexts, companyID)

	log.Printf("[EmailWorkerManager] Stopped worker for company %s", companyID)
}

// stopAllWorkers stops all running workers
func (m *EmailWorkerManager) stopAllWorkers() {
	m.mu.Lock()
	for companyID, cancel := range m.workerContexts {
		cancel()
		log.Printf("[EmailWorkerManager] Stopped worker for company %s", companyID)
	}

	m.workers = make(map[string]*CompanyEmailWorker)
	m.workerContexts = make(map[string]context.CancelFunc)
	m.mu.Unlock()

	// Wait for all workers to finish with timeout
	log.Println("[EmailWorkerManager] Waiting for all workers to finish...")

	c := make(chan struct{})
	go func() {
		defer close(c)
		m.wg.Wait()
	}()

	select {
	case <-c:
		log.Println("[EmailWorkerManager] All workers stopped successfully")
	case <-time.After(5 * time.Second): // 5 second timeout
		log.Println("[EmailWorkerManager] Timeout waiting for workers to stop")
	}
}
