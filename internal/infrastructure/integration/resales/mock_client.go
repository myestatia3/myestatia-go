package resales

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

// Client defines the interface for interacting with Resales Online API
type Client interface {
	TestConnection(ctx context.Context, apiKey string, agencyID string) error
	GetProperties(ctx context.Context, apiKey string, agencyID string, page int, pageSize int) (*APIResponse, error)
}

// MockClient is a mock implementation of the Resales Client
type MockClient struct {
	// Add potential state here if needed for tests
}

func NewMockClient() *MockClient {
	return &MockClient{}
}

func (m *MockClient) TestConnection(ctx context.Context, apiKey string, agencyID string) error {
	// Simulate latency
	time.Sleep(500 * time.Millisecond)

	if apiKey == "error" {
		return fmt.Errorf("invalid api key")
	}
	return nil
}

func (m *MockClient) GetProperties(ctx context.Context, apiKey string, agencyID string, page int, pageSize int) (*APIResponse, error) {
	// Simulate latency
	time.Sleep(500 * time.Millisecond)

	if apiKey == "error" {
		return nil, fmt.Errorf("authentication failed")
	}

	// Generate mock properties
	props := []Property{}
	for i := 0; i < pageSize; i++ {
		ref := fmt.Sprintf("R%d", rand.Intn(100000)+3000000)
		props = append(props, Property{
			Reference:   ref,
			Price:       rand.Intn(2000000) + 200000,
			Currency:    "EUR",
			Description: "Beautiful property in Costa del Sol with amazing sea views...",
			Location:    "Marbella",
			Type:        "Apartment",
			Bedrooms:    rand.Intn(4) + 1,
			Bathrooms:   rand.Intn(3) + 1,
			Built:       rand.Intn(200) + 80,
			MainImage:   fmt.Sprintf("https://picsum.photos/seed/%s/800/600", ref),
			PropertyFeatures: PropertyFeatures{
				Category: []FeatureCategory{
					{Type: "Setting", Value: []string{"Close To Golf", "Close To Sea"}},
					{Type: "Pool", Value: []string{"Communal"}},
				},
			},
		})
	}

	return &APIResponse{
		Transaction: Transaction{Status: "success", Version: "6.0"},
		QueryInfo: &QueryInfo{
			QueryId:       "mock-query-id",
			PropertyCount: 100,
			CurrentPage:   page,
		},
		Property: props,
	}, nil
}
