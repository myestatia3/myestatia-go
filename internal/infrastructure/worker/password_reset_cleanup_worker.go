package worker

import (
	"context"
	"log"
	"time"

	"github.com/myestatia/myestatia-go/internal/infrastructure/repository"
)

// PasswordResetCleanupWorker handles periodic cleanup of expired password reset tokens
type PasswordResetCleanupWorker struct {
	passwordResetRepo repository.PasswordResetRepository
	interval          time.Duration
}

// NewPasswordResetCleanupWorker creates a new cleanup worker
func NewPasswordResetCleanupWorker(passwordResetRepo repository.PasswordResetRepository) *PasswordResetCleanupWorker {
	return &PasswordResetCleanupWorker{
		passwordResetRepo: passwordResetRepo,
		interval:          2 * time.Hour, // Run every 2 hours
	}
}

// Start begins the cleanup worker
func (w *PasswordResetCleanupWorker) Start(ctx context.Context) {
	log.Printf("[PasswordResetCleanup] Worker started, running every %v", w.interval)

	// Run immediately on start
	w.cleanup(ctx)

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("[PasswordResetCleanup] Worker stopped")
			return
		case <-ticker.C:
			w.cleanup(ctx)
		}
	}
}

// cleanup performs the actual cleanup operation
func (w *PasswordResetCleanupWorker) cleanup(ctx context.Context) {
	log.Println("[PasswordResetCleanup] Starting cleanup of expired tokens...")

	if err := w.passwordResetRepo.DeleteExpired(ctx); err != nil {
		log.Printf("[PasswordResetCleanup] ERROR: Failed to delete expired tokens: %v", err)
		return
	}

	log.Println("[PasswordResetCleanup] ✓ Cleanup completed successfully")
}
