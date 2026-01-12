package mocks

import (
	"context"
	"mime/multipart"

	"github.com/stretchr/testify/mock"
)

type StorageServiceMock struct {
	mock.Mock
}

func (m *StorageServiceMock) UploadFile(ctx context.Context, file multipart.File, header *multipart.FileHeader) (string, error) {
	args := m.Called(ctx, file, header)
	return args.String(0), args.Error(1)
}
