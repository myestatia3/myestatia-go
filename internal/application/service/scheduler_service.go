package service

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
)

// SyncConfig stores sync configuration for a company
type SyncConfig struct {
	CompanyID string
	APIKey    string
	AgencyID  string
	Enabled   bool
}

// SchedulerService handles scheduled tasks like nightly syncs
type SchedulerService struct {
	cron               *cron.Cron
	integrationService *IntegrationService
	syncConfigs        map[string]*SyncConfig // key: companyID
	configMutex        sync.RWMutex
}

// NewSchedulerService creates a new scheduler service
func NewSchedulerService(integrationService *IntegrationService) *SchedulerService {
	return &SchedulerService{
		cron:               cron.New(),
		integrationService: integrationService,
		syncConfigs:        make(map[string]*SyncConfig),
	}
}

// Start begins the scheduler
func (s *SchedulerService) Start() {
	// Schedule nightly sync at 3:00 AM
	_, err := s.cron.AddFunc("0 3 * * *", s.runNightlySync)
	if err != nil {
		log.Printf("[SCHEDULER] Failed to schedule nightly sync: %v", err)
		return
	}

	log.Printf("[SCHEDULER] Nightly sync scheduled for 3:00 AM")
	s.cron.Start()
}

// Stop gracefully stops the scheduler
func (s *SchedulerService) Stop() {
	ctx := s.cron.Stop()
	<-ctx.Done()
	log.Printf("[SCHEDULER] Scheduler stopped")
}

// RegisterCompanySync registers a company for nightly sync
func (s *SchedulerService) RegisterCompanySync(companyID, apiKey, agencyID string) {
	s.configMutex.Lock()
	defer s.configMutex.Unlock()

	s.syncConfigs[companyID] = &SyncConfig{
		CompanyID: companyID,
		APIKey:    apiKey,
		AgencyID:  agencyID,
		Enabled:   true,
	}
	log.Printf("[SCHEDULER] Company %s registered for nightly sync", companyID)
}

// UnregisterCompanySync removes a company from nightly sync
func (s *SchedulerService) UnregisterCompanySync(companyID string) {
	s.configMutex.Lock()
	defer s.configMutex.Unlock()

	delete(s.syncConfigs, companyID)
	log.Printf("[SCHEDULER] Company %s unregistered from nightly sync", companyID)
}

// SetSyncEnabled enables or disables sync for a company
func (s *SchedulerService) SetSyncEnabled(companyID string, enabled bool) {
	s.configMutex.Lock()
	defer s.configMutex.Unlock()

	if config, exists := s.syncConfigs[companyID]; exists {
		config.Enabled = enabled
		log.Printf("[SCHEDULER] Company %s sync enabled: %v", companyID, enabled)
	}
}

// GetSyncConfig returns the sync configuration for a company
func (s *SchedulerService) GetSyncConfig(companyID string) *SyncConfig {
	s.configMutex.RLock()
	defer s.configMutex.RUnlock()
	return s.syncConfigs[companyID]
}

// GetAllSyncConfigs returns all registered sync configurations
func (s *SchedulerService) GetAllSyncConfigs() []*SyncConfig {
	s.configMutex.RLock()
	defer s.configMutex.RUnlock()

	configs := make([]*SyncConfig, 0, len(s.syncConfigs))
	for _, config := range s.syncConfigs {
		configs = append(configs, config)
	}
	return configs
}

// runNightlySync executes the nightly sync for all registered companies
func (s *SchedulerService) runNightlySync() {
	s.configMutex.RLock()
	configs := make([]*SyncConfig, 0, len(s.syncConfigs))
	for _, config := range s.syncConfigs {
		if config.Enabled {
			configs = append(configs, config)
		}
	}
	s.configMutex.RUnlock()

	if len(configs) == 0 {
		log.Printf("[SCHEDULER] No companies registered for nightly sync")
		return
	}

	log.Printf("[SCHEDULER] Starting nightly sync for %d companies at %s", len(configs), time.Now().Format(time.RFC3339))

	for _, config := range configs {
		go s.syncCompany(config)
	}
}

// syncCompany runs the sync for a single company
func (s *SchedulerService) syncCompany(config *SyncConfig) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	settings := map[string]string{
		"api_key":   config.APIKey,
		"agency_id": config.AgencyID,
	}

	log.Printf("[SCHEDULER] Starting sync for company %s", config.CompanyID)

	result, err := s.integrationService.SyncAllProperties(ctx, config.CompanyID, settings)
	if err != nil {
		log.Printf("[SCHEDULER] Sync failed for company %s: %v", config.CompanyID, err)
		return
	}

	log.Printf("[SCHEDULER] Sync completed for company %s: %d synced, %d created, %d updated, %d errors",
		config.CompanyID, result.PropertiesSynced, result.CreatedCount, result.UpdatedCount, result.ErrorCount)
}

// TriggerManualSync runs a sync immediately for a specific company
func (s *SchedulerService) TriggerManualSync(ctx context.Context, companyID string) (*SyncResult, error) {
	config := s.GetSyncConfig(companyID)
	if config == nil {
		return nil, nil
	}

	settings := map[string]string{
		"api_key":   config.APIKey,
		"agency_id": config.AgencyID,
	}

	return s.integrationService.SyncAllProperties(ctx, companyID, settings)
}
