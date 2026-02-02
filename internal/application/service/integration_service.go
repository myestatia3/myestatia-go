package service

import (
	"context"
	"errors"

	"github.com/myestatia/myestatia-go/internal/domain/entity"
	resalesIntegration "github.com/myestatia/myestatia-go/internal/infrastructure/integration/resales"
)

type IntegrationService struct {
	// In the future this might use a repository to save configs
	// For now we might just hold config in memory or assume passed from frontend for tests
	resalesClient resalesIntegration.Client
}

func NewIntegrationService() *IntegrationService {
	return &IntegrationService{
		resalesClient: resalesIntegration.NewMockClient(), // Injecting Mock for now
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

// SyncProperties fetches properties from the integration and mapped them
// This is a "dry run" or "preview" in this step, or actual sync
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
	for _, rProp := range resp.Property {
		// Mock Company ID for preview
		domainProp, err := resalesIntegration.MapToDomain(rProp, "preview-company-id")
		if err == nil {
			properties = append(properties, *domainProp)
		}
	}

	return properties, nil
}
