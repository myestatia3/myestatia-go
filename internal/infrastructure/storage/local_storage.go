package storage

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/myestatia/myestatia-go/internal/application/service"
)

type LocalStorageService struct {
	BaseDir string
	BaseURL string
}

func NewLocalStorageService(baseDir, baseURL string) service.StorageService {
	// Ensure directory exists
	if _, err := os.Stat(baseDir); os.IsNotExist(err) {
		_ = os.MkdirAll(baseDir, 0755)
	}
	return &LocalStorageService{
		BaseDir: baseDir,
		BaseURL: baseURL,
	}
}

func (s *LocalStorageService) UploadFile(ctx context.Context, file multipart.File, header *multipart.FileHeader) (string, error) {
	// Generate unique filename
	ext := filepath.Ext(header.Filename)
	uniqueName := fmt.Sprintf("%s_%s%s",
		time.Now().Format("20060102"),
		uuid.New().String()[:8],
		ext,
	)

	// Create destination file
	dstPath := filepath.Join(s.BaseDir, uniqueName)
	dst, err := os.Create(dstPath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close()

	// Copy content
	if _, err := io.Copy(dst, file); err != nil {
		return "", fmt.Errorf("failed to save file: %w", err)
	}

	// Return public URL
	return fmt.Sprintf("%s/%s", s.BaseURL, uniqueName), nil
}
