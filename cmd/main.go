package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	router "github.com/myestatia/myestatia-go/internal/adapters/input"
	handlers "github.com/myestatia/myestatia-go/internal/adapters/input/handler"
	"github.com/myestatia/myestatia-go/internal/adapters/input/middleware"
	"github.com/myestatia/myestatia-go/internal/application/service"
	entity "github.com/myestatia/myestatia-go/internal/domain/entity"
	database "github.com/myestatia/myestatia-go/internal/infrastructure/database"
	"github.com/myestatia/myestatia-go/internal/infrastructure/email"
	googleoauth "github.com/myestatia/myestatia-go/internal/infrastructure/oauth2"
	repository "github.com/myestatia/myestatia-go/internal/infrastructure/repository"
	"github.com/myestatia/myestatia-go/internal/infrastructure/seed"
	"github.com/myestatia/myestatia-go/internal/infrastructure/storage"
	"github.com/myestatia/myestatia-go/internal/infrastructure/worker"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	cfg := database.LoadConfig()

	if cfg.Host == "" {
		cfg = database.Config{
			Host:     "localhost",
			User:     "user",
			Password: "pass",
			DBName:   "mydb",
			Port:     5432,
			SSLMode:  "disable",
		}
	}

	db := database.InitDB(cfg)

	err := db.AutoMigrate(
		&entity.Lead{},
		&entity.Message{},
		&entity.Summary{},
		&entity.PropertySubtype{},
		&entity.Property{},
		&entity.SystemConfig{},
		&entity.Agent{},
		&entity.Company{},
		&entity.CompanyEmailConfig{},
		&entity.ProcessedEmail{},
		&entity.PasswordReset{},
	)
	if err != nil {
		log.Fatalf("Error migrating database: %v", err)
	}

	log.Println("Database migrated successfully")

	if err := seed.SeedPropertySubtypes(db); err != nil {
		log.Printf("Error seeding subtypes: %v", err)
	}

	leadRepo := repository.NewLeadRepository(db)
	leadSvc := service.NewLeadService(leadRepo)
	leadHandler := handlers.NewLeadHandler(leadSvc)

	propertyRepo := repository.NewPropertyRepository(db)
	propertyService := service.NewPropertyService(propertyRepo)

	companyRepo := repository.NewCompanyRepository(db)
	companyService := service.NewCompanyService(companyRepo)
	companyHandler := handlers.NewCompanyHandler(companyService)

	agentRepo := repository.NewAgentRepository(db)
	agentService := service.NewAgentService(agentRepo, propertyRepo)
	agentHandler := handlers.NewAgentHandler(agentService)

	// Storage
	storageService := storage.NewLocalStorageService("uploads", "http://localhost:8080/uploads")
	propertyHandler := handlers.NewPropertyHandler(propertyService, agentService, companyService, storageService)

	messageRepo := repository.NewMessageRepository(db)
	messageService := service.NewMessageService(messageRepo)
	messageHandler := handlers.NewMessageHandler(messageService)

	// Auth handler will be initialized later with invitation service

	// Email Configuration Service and Manager
	encryptionKey := os.Getenv("ENCRYPTION_KEY")
	if len(encryptionKey) != 32 {
		log.Fatalf("ENCRYPTION_KEY must be exactly 32 bytes, got %d bytes", len(encryptionKey))
	}

	emailConfigRepo := repository.NewCompanyEmailConfigRepository(db)
	emailConfigService := service.NewCompanyEmailConfigService(emailConfigRepo, encryptionKey)
	emailConfigHandler := handlers.NewCompanyEmailConfigHandler(emailConfigService)

	// Repositories for worker
	processedEmailRepo := repository.NewProcessedEmailRepository(db)
	passwordResetRepo := repository.NewPasswordResetRepository(db)

	// Start Email Worker Manager (multi-company support)
	emailWorkerManager := worker.NewEmailWorkerManager(
		emailConfigService,
		propertyRepo,
		leadRepo,
		processedEmailRepo,
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go emailWorkerManager.Start(ctx)

	// Start Password Reset Cleanup Worker
	passwordResetCleanupWorker := worker.NewPasswordResetCleanupWorker(passwordResetRepo)
	go passwordResetCleanupWorker.Start(ctx)

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Google OAuth2 Configuration
	googleClientID := os.Getenv("GOOGLE_CLIENT_ID")
	googleClientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	googleRedirectURL := os.Getenv("GOOGLE_REDIRECT_URL")

	if googleRedirectURL == "" {
		googleRedirectURL = "http://localhost:8080/api/v1/auth/google/callback" // Default for development
	}

	if googleClientID == "" || googleClientSecret == "" {
		log.Println("WARNING: GOOGLE_CLIENT_ID or GOOGLE_CLIENT_SECRET not set. OAuth2 will not work.")
		log.Println("See GOOGLE_OAUTH_SETUP.md for instructions")
	}

	oauth2Config := googleoauth.NewOAuth2Config(googleClientID, googleClientSecret, googleRedirectURL)
	googleOAuthHandler := handlers.NewGoogleOAuthHandler(oauth2Config, emailConfigService, encryptionKey)

	// Password Reset Setup - Try Resend first, fallback to SMTP
	resendConfig := email.LoadResendConfig()
	smtpConfig := email.LoadSMTPConfig()

	var passwordResetEmailSender interface {
		SendPasswordResetEmail(to, token string) error
	}

	if resendConfig.IsValid() {
		log.Println("✓ Using Resend for password reset emails")
		passwordResetEmailSender = email.NewResendEmailSender(resendConfig)
	} else {
		log.Println("Resend not configured, trying SMTP...")
		if smtpConfig.IsValid() {
			log.Println("✓ Using SMTP for password reset emails")
			passwordResetEmailSender = email.NewEmailSender(smtpConfig)
		} else {
			log.Println("WARNING: Neither Resend nor SMTP configured. Password reset emails will not be sent.")
			log.Println("For Resend (recommended): Set RESEND_API_KEY and RESEND_FROM_EMAIL")
			log.Println("For SMTP: Set SMTP_HOST, SMTP_PORT, SMTP_USERNAME, SMTP_PASSWORD, and SMTP_FROM")
			// Create a dummy sender that will fail gracefully
			passwordResetEmailSender = email.NewEmailSender(smtpConfig)
		}
	}

	passwordResetHandler := handlers.NewPasswordResetHandler(agentService, passwordResetRepo, passwordResetEmailSender)

	// Presentation Service
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "default-secret-key-change-me"
	}
	presentationService := service.NewPresentationService(leadRepo, propertyRepo, agentRepo, companyRepo, jwtSecret)
	presentationHandler := handlers.NewPresentationHandler(presentationService)

	// Initialize auth handler (invitation service removed)
	authHandler := handlers.NewAuthHandler(agentService, companyService)

	// Integration
	integrationService := service.NewIntegrationService()
	integrationHandler := handlers.NewIntegrationHandler(integrationService)

	mux := router.NewRouter(leadHandler, propertyHandler, companyHandler, agentHandler, messageHandler, authHandler, emailConfigHandler, googleOAuthHandler, passwordResetHandler, presentationHandler, integrationHandler)

	// Wrap the router with CORS middleware
	// Add static file handler for uploads
	fileServer := http.StripPrefix("/uploads", http.FileServer(http.Dir("uploads")))

	finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "" && len(r.URL.Path) > 8 && r.URL.Path[:8] == "/uploads" {
			fileServer.ServeHTTP(w, r)
			return
		}
		middleware.CorsMiddleware(mux).ServeHTTP(w, r)
	})

	// Create HTTP server
	server := &http.Server{
		Addr:    ":8080",
		Handler: finalHandler,
	}

	// Start server in goroutine
	go func() {
		log.Println("listening on :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	<-sigChan
	log.Println("Received shutdown signal, stopping services...")

	// Cancel worker manager context
	cancel()

	// Shutdown HTTP server with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	log.Println("Server stopped gracefully")

	// Force exit to ensure no hanging goroutines keep the process alive
	os.Exit(0)
}
