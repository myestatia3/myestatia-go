package service

import (
	"context"
	"errors"

	"github.com/google/uuid"

	entity "github.com/myestatia/myestatia-go/internal/domain/entity"
	repository "github.com/myestatia/myestatia-go/internal/infrastructure/repository"
)

type PropertyService struct {
	repo repository.PropertyRepository
}

func NewPropertyService(repo repository.PropertyRepository) *PropertyService {
	return &PropertyService{repo: repo}
}

func (s *PropertyService) CreateProperty(ctx context.Context, p *entity.Property) (*entity.Property, bool, error) {
	if p.Reference == "" || p.CompanyID == "" {
		return nil, false, errors.New("reference and company_id are required")
	}

	existing, err := s.repo.FindByReference(ctx, p.Reference)
	if err != nil {
		return nil, false, err
	}
	if existing != nil {
		return existing, false, nil // Ya existe
	}

	if p.ID == "" {
		p.ID = uuid.New().String()
	}

	if err := s.repo.Create(p); err != nil {
		return nil, false, err
	}

	return p, true, nil
}

func (s *PropertyService) GetPropertyByID(ctx context.Context, id string) (*entity.Property, error) {
	return s.repo.FindByID(id)
}

func (s *PropertyService) GetAllProperties(ctx context.Context) ([]entity.Property, error) {
	return s.repo.FindAll()
}

func (s *PropertyService) UpdateProperty(ctx context.Context, p *entity.Property) error {
	return s.repo.Update(p)
}

func (s *PropertyService) DeleteProperty(ctx context.Context, id string) error {
	return s.repo.Delete(id)
}

func (s *PropertyService) FindAllByCompanyID(ctx context.Context, companyID string) ([]entity.Property, error) {
	return s.repo.FindAllByCompanyID(ctx, companyID)
}

func (s *PropertyService) SearchProperties(ctx context.Context, filter entity.PropertyFilter) ([]entity.Property, error) {
	return s.repo.Search(ctx, filter)
}

func (s *PropertyService) ListSubtypes(ctx context.Context, propertyType string) ([]entity.PropertySubtype, error) {
	return s.repo.FindSubtypes(ctx, propertyType)
}
