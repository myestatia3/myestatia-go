package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/myestatia/myestatia-go/internal/domain/entity"
	resalesIntegration "github.com/myestatia/myestatia-go/internal/infrastructure/integration/resales"
	"github.com/myestatia/myestatia-go/internal/infrastructure/repository"
	"gorm.io/gorm"
)

type PropertyRepository interface {
	FindByReference(ctx context.Context, reference string) (*entity.Property, error)
	Create(property *entity.Property) error
	Update(property *entity.Property) error
}

// SyncStatus tracks the progress of an ongoing synchronization
type SyncStatus struct {
	Status          string     `json:"status"` // "running", "completed", "failed"
	TotalProperties int        `json:"total_properties"`
	Processed       int        `json:"processed"`
	Created         int        `json:"created"`
	Updated         int        `json:"updated"`
	Errors          int        `json:"errors"`
	StartedAt       time.Time  `json:"started_at"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
	ErrorMessage    string     `json:"error_message,omitempty"`
	CurrentPage     int        `json:"current_page"`
	TotalPages      int        `json:"total_pages"`
}

type IntegrationService struct {
	resalesClient resalesIntegration.Client
	propertyRepo  PropertyRepository
	agentRepo     repository.AgentRepository
	syncStatuses  map[string]*SyncStatus // key: companyID
	syncMutex     sync.RWMutex
}

// NewIntegrationService creates the service with real client
func NewIntegrationService(propertyRepo PropertyRepository, agentRepo repository.AgentRepository) *IntegrationService {
	return &IntegrationService{
		resalesClient: resalesIntegration.NewRealClient(),
		propertyRepo:  propertyRepo,
		agentRepo:     agentRepo,
		syncStatuses:  make(map[string]*SyncStatus),
	}
}

// TestConnection verifies if the credentials are valid
func (s *IntegrationService) TestConnection(ctx context.Context, integrationID string, settings map[string]string) error {
	if integrationID == "resales" {
		apiKey := settings["api_key"]
		agencyID := settings["agency_id"]
		if apiKey == "" {
			return errors.New("api_key is required")
		}
		return s.resalesClient.TestConnection(ctx, apiKey, agencyID)
	}
	return nil // Other integrations not implemented or always succeed mock
}

// PreviewProperties fetches a small number of properties for preview
func (s *IntegrationService) PreviewProperties(ctx context.Context, integrationID string, settings map[string]string) ([]entity.Property, error) {
	if integrationID != "resales" {
		return nil, errors.New("integration not supported for sync")
	}

	apiKey := settings["api_key"]
	agencyID := settings["agency_id"]

	resp, err := s.resalesClient.GetProperties(ctx, apiKey, agencyID, 1, 5) // Fetch first page/5 items for preview
	if err != nil {
		return nil, err
	}

	var properties []entity.Property
	for i, rProp := range resp.Property {
		// Mock Company ID for preview
		domainProp, err := resalesIntegration.MapToDomain(rProp, "preview-company-id", i)
		if err == nil {
			properties = append(properties, *domainProp)
		}
	}

	return properties, nil
}

// SyncResult contains the result of a synchronization operation
type SyncResult struct {
	PropertiesSynced int       `json:"properties_synced"`
	CreatedCount     int       `json:"created_count"`
	UpdatedCount     int       `json:"updated_count"`
	ErrorCount       int       `json:"error_count"`
	LastSync         time.Time `json:"last_sync"`
}

// SyncAllProperties performs a full synchronization with the Resales API
func (s *IntegrationService) SyncAllProperties(ctx context.Context, companyID string, settings map[string]string) (*SyncResult, error) {
	apiKey := settings["api_key"]
	agencyID := settings["agency_id"]

	if apiKey == "" || agencyID == "" {
		return nil, errors.New("api_key and agency_id are required")
	}

	// Find admin agent for this company to assign as createdByAgentId
	var adminAgentID *string
	if s.agentRepo != nil {
		admin, err := s.agentRepo.FindAdminByCompanyID(companyID)
		if err != nil {
			log.Printf("[RESALES] Warning: could not find admin agent for company %s: %v", companyID, err)
		} else if admin != nil {
			adminAgentID = &admin.ID
			log.Printf("[RESALES] Using admin agent %s (%s) for imported properties", admin.Name, admin.ID)
		}
	}

	log.Printf("[RESALES] Starting full sync for company %s", companyID)

	// Initialize sync status
	now := time.Now()
	s.syncMutex.Lock()
	s.syncStatuses[companyID] = &SyncStatus{
		Status:    "running",
		StartedAt: now,
	}
	s.syncMutex.Unlock()

	// Fetch first page to get total count
	page1, err := s.resalesClient.GetProperties(ctx, apiKey, agencyID, 1, 50)
	if err != nil {
		s.updateSyncStatus(companyID, func(status *SyncStatus) {
			status.Status = "failed"
			status.ErrorMessage = fmt.Sprintf("Failed to fetch first page: %v", err)
			completedAt := time.Now()
			status.CompletedAt = &completedAt
		})
		return nil, fmt.Errorf("failed to fetch first page: %w", err)
	}

	totalProperties := page1.QueryInfo.PropertyCount
	totalPages := (totalProperties / 50) + 1
	log.Printf("[RESALES] Total properties to sync: %d", totalProperties)

	// Update status with totals
	s.updateSyncStatus(companyID, func(status *SyncStatus) {
		status.TotalProperties = totalProperties
		status.TotalPages = totalPages
		status.CurrentPage = 1
	})

	importedCount := 0
	createdCount := 0
	updatedCount := 0
	errorCount := 0

	// CONCURRENT PROCESSING with worker pool
	const maxWorkers = 10 // Process 10 properties concurrently
	semaphore := make(chan struct{}, maxWorkers)
	var wg sync.WaitGroup
	var mu sync.Mutex // Protect shared counters

	processProperty := func(apiProp resalesIntegration.Property) {
		defer wg.Done()
		defer func() { <-semaphore }() // Release slot

		mu.Lock()
		currentProcessed := importedCount
		mu.Unlock()

		created, err := s.upsertProperty(ctx, apiProp, companyID, currentProcessed, adminAgentID)

		mu.Lock()
		if err != nil {
			log.Printf("[RESALES] Error upserting property %s: %v", apiProp.AgencyRef, err)
			errorCount++
		} else {
			if created {
				createdCount++
			} else {
				updatedCount++
			}
			importedCount++
		}

		// Update progress
		s.updateSyncStatus(companyID, func(status *SyncStatus) {
			status.Processed = importedCount
			status.Created = createdCount
			status.Updated = updatedCount
			status.Errors = errorCount
		})
		mu.Unlock()
	}

	// Process first page concurrently
	for _, apiProp := range page1.Property {
		wg.Add(1)
		semaphore <- struct{}{} // Acquire slot
		go processProperty(apiProp)
	}

	// Fetch remaining pages
	for page := 2; page <= totalPages; page++ {
		log.Printf("[RESALES] Fetching page %d/%d", page, totalPages)

		// Update current page
		s.updateSyncStatus(companyID, func(status *SyncStatus) {
			status.CurrentPage = page
		})

		pageData, err := s.resalesClient.GetProperties(ctx, apiKey, agencyID, page, 50)
		if err != nil {
			log.Printf("[RESALES] Error fetching page %d: %v", page, err)
			mu.Lock()
			errorCount++
			s.updateSyncStatus(companyID, func(status *SyncStatus) {
				status.Errors = errorCount
			})
			mu.Unlock()
			continue
		}

		// Process properties from this page concurrently
		for _, apiProp := range pageData.Property {
			wg.Add(1)
			semaphore <- struct{}{} // Acquire slot
			go processProperty(apiProp)
		}
	}

	// Wait for all workers to complete
	wg.Wait()
	log.Printf("[RESALES] All properties processed: %d created, %d updated, %d errors", createdCount, updatedCount, errorCount)

	completedAt := time.Now()
	log.Printf("[RESALES] Sync completed: %d created, %d updated, %d errors", createdCount, updatedCount, errorCount)

	// Mark as completed
	s.updateSyncStatus(companyID, func(status *SyncStatus) {
		status.Status = "completed"
		status.CompletedAt = &completedAt
	})

	return &SyncResult{
		PropertiesSynced: importedCount,
		CreatedCount:     createdCount,
		UpdatedCount:     updatedCount,
		ErrorCount:       errorCount,
		LastSync:         completedAt,
	}, nil
}

// upsertProperty creates or updates a single property based on Reference
// processedCount is used to limit image downloads to first N properties
// adminAgentID is assigned as CreatedByAgentID for newly created properties
func (s *IntegrationService) upsertProperty(ctx context.Context, apiProp resalesIntegration.Property, companyID string, processedCount int, adminAgentID *string) (created bool, err error) {
	// Map Resales property to domain entity with processedCount for image limiting
	domainProp, err := resalesIntegration.MapToDomain(apiProp, companyID, processedCount)
	if err != nil {
		return false, fmt.Errorf("failed to map property: %w", err)
	}

	// Set agent assignment for new properties
	if adminAgentID != nil {
		domainProp.CreatedByAgentID = adminAgentID
	}

	// Try to find existing property by Reference (AgencyRef)
	existing, err := s.propertyRepo.FindByReference(ctx, domainProp.Reference)

	if err != nil && err != gorm.ErrRecordNotFound {
		return false, fmt.Errorf("failed to query property: %w", err)
	}

	if err == gorm.ErrRecordNotFound {
		// CREATE NEW PROPERTY
		if err := s.propertyRepo.Create(domainProp); err != nil {
			return false, fmt.Errorf("failed to create property: %w", err)
		}
		return true, nil
	}

	// UPDATE EXISTING PROPERTY
	// Preserve certain fields from existing record
	domainProp.ID = existing.ID
	domainProp.CreatedAt = existing.CreatedAt
	// Only preserve existing creator if it was set, otherwise use the admin agent
	if existing.CreatedByAgentID != nil && *existing.CreatedByAgentID != "" {
		domainProp.CreatedByAgentID = existing.CreatedByAgentID
	}
	// If neither existing nor new has an agent, domainProp.CreatedByAgentID may be nil or the admin (set above)

	if err := s.propertyRepo.Update(domainProp); err != nil {
		return false, fmt.Errorf("failed to update property: %w", err)
	}

	return false, nil
}

// GetSyncStatus returns the current sync status for a company
func (s *IntegrationService) GetSyncStatus(companyID string) *SyncStatus {
	s.syncMutex.RLock()
	defer s.syncMutex.RUnlock()

	status, exists := s.syncStatuses[companyID]
	if !exists {
		return nil
	}

	// Return a copy to avoid race conditions
	statusCopy := *status
	return &statusCopy
}

// updateSyncStatus updates the sync status for a company
func (s *IntegrationService) updateSyncStatus(companyID string, updater func(*SyncStatus)) {
	s.syncMutex.Lock()
	defer s.syncMutex.Unlock()

	if status, exists := s.syncStatuses[companyID]; exists {
		updater(status)
	}
}
