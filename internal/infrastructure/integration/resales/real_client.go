package resales

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// RealClient implements the actual HTTP client for Resales API
type RealClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewRealClient creates a new real HTTP client
func NewRealClient() *RealClient {
	return &RealClient{
		baseURL: "https://webapi.resales-online.com/V6",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// TestConnection verifies if the credentials are valid
func (c *RealClient) TestConnection(ctx context.Context, apiKey string, agencyID string) error {
	// Build URL with auth params - minimal request to test connection
	reqURL := fmt.Sprintf("%s/SearchProperties?p1=%s&p2=%s&p_agency_filterid=1&P_PageSize=1",
		c.baseURL, agencyID, apiKey)

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to call API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API returned status code: %d", resp.StatusCode)
	}

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if apiResp.Transaction.Status != "success" {
		return fmt.Errorf("API returned status: %s", apiResp.Transaction.Status)
	}

	return nil
}

// GetProperties fetches properties from the API with pagination support
func (c *RealClient) GetProperties(ctx context.Context, apiKey string, agencyID string, page int, pageSize int) (*APIResponse, error) {
	// Build URL with parameters
	reqURL := fmt.Sprintf("%s/SearchProperties?p1=%s&p2=%s&p_agency_filterid=1&P_PageSize=%d",
		c.baseURL, agencyID, apiKey, pageSize)

	// Add page number if not first page
	if page > 1 {
		reqURL += fmt.Sprintf("&P_PageNo=%d", page)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if apiResp.Transaction.Status != "success" {
		return nil, fmt.Errorf("API error: %s", apiResp.Transaction.Status)
	}

	return &apiResp, nil
}

// downloadImages downloads all images for a property and returns local paths
// Falls back to external URLs if download fails
func downloadImages(prop Property, reference string) (string, []string, error) {
	var mainImage string
	var photos []string

	// Create upload directory
	uploadDir := "./uploads/properties"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		log.Printf("[RESALES] Failed to create upload directory: %v", err)
		// Fallback to external URLs
		if prop.MainImage != "" {
			mainImage = prop.MainImage
			photos = append(photos, prop.MainImage)
		}
		for _, pic := range prop.Pictures.Picture {
			if pic.URL != "" && pic.URL != prop.MainImage {
				photos = append(photos, pic.URL)
			}
		}
		return mainImage, photos, err
	}

	// Download main image
	if prop.MainImage != "" {
		localPath, err := downloadSingleImage(prop.MainImage, reference, 0, uploadDir)
		if err == nil {
			mainImage = localPath
			photos = append(photos, localPath)
		} else {
			log.Printf("[RESALES] Failed to download main image for %s: %v, using CDN URL", reference, err)
			mainImage = prop.MainImage
			photos = append(photos, prop.MainImage)
		}
	}

	// Download additional pictures
	for i, pic := range prop.Pictures.Picture {
		if pic.URL == "" || pic.URL == prop.MainImage {
			continue
		}

		localPath, err := downloadSingleImage(pic.URL, reference, i+1, uploadDir)
		if err == nil {
			photos = append(photos, localPath)
		} else {
			log.Printf("[RESALES] Failed to download image %d for %s: %v, using CDN URL", i+1, reference, err)
			photos = append(photos, pic.URL)
		}
	}

	return mainImage, photos, nil
}

// downloadSingleImage downloads a single image and returns the local path
func downloadSingleImage(imageURL, reference string, index int, uploadDir string) (string, error) {
	// Sanitize reference to avoid filesystem issues (e.g., "/" in AgencyRef like "7272-00444/7272")
	safeReference := sanitizeFilename(reference)

	// Determine file extension
	ext := filepath.Ext(imageURL)
	if ext == "" || len(ext) > 5 { // Sometimes URL params make ext too long
		ext = ".jpg"
	}

	// Create filename: {AgencyRef}-{index}.jpg
	filename := fmt.Sprintf("%s-%d%s", safeReference, index, ext)
	localPath := filepath.Join(uploadDir, filename)

	// Check if already exists
	if _, err := os.Stat(localPath); err == nil {
		// File exists, return existing path
		return "/uploads/properties/" + filename, nil
	}

	// Download image with timeout
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(imageURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	// Create file
	file, err := os.Create(localPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Copy data
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		os.Remove(localPath) // Cleanup on failure
		return "", err
	}

	return "/uploads/properties/" + filename, nil
}

// sanitizeFilename replaces invalid filesystem characters with underscores
func sanitizeFilename(s string) string {
	// Replace common invalid characters for Windows/Unix filesystems
	replacer := map[rune]rune{
		'/':  '_',
		'\\': '_',
		':':  '_',
		'*':  '_',
		'?':  '_',
		'"':  '_',
		'<':  '_',
		'>':  '_',
		'|':  '_',
	}

	result := make([]rune, 0, len(s))
	for _, char := range s {
		if replacement, ok := replacer[char]; ok {
			result = append(result, replacement)
		} else {
			result = append(result, char)
		}
	}
	return string(result)
}
