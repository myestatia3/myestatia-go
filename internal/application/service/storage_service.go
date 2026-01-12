package service

import (
	"context"
	"mime/multipart"
)

type StorageService interface {
	UploadFile(ctx context.Context, file multipart.File, header *multipart.FileHeader) (string, error)
}
