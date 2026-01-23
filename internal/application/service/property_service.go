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

func (s *PropertyService) UpdateProperty(ctx context.Context, p *entity.Property, executorID, executorRole string) error {
	existing, err := s.repo.FindByID(p.ID)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("property not found")
	}

	// Permission check: Admin or Creator
	isCreator := existing.CreatedByAgentID != nil && *existing.CreatedByAgentID == executorID
	isAdmin := executorRole == "admin"

	if !isCreator && !isAdmin {
		return errors.New("unauthorized: only admin or creator can update this property")
	}

	// Tenemos que adaptar esto para que funcione con el update parcial correctamente
	if p.Status != "" {
		existing.Status = p.Status
	}
	if p.Title != "" {
		existing.Title = p.Title
	}
	if p.Description != "" {
		existing.Description = p.Description
	}
	if p.Price != 0 {
		existing.Price = p.Price
	}
	// Critical fields that must NOT be cleared by partial update:
	// CompanyID, Reference, CreatedByAgentID -> already in 'existing'

	// Use existing as the object to save
	return s.repo.Update(existing)
}

func (s *PropertyService) DeleteProperty(ctx context.Context, id string, executorID, executorRole string) error {
	existing, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("property not found")
	}

	// Permission check: Admin or Creator
	isCreator := existing.CreatedByAgentID != nil && *existing.CreatedByAgentID == executorID
	isAdmin := executorRole == "admin"

	if !isCreator && !isAdmin {
		return errors.New("unauthorized: only admin or creator can delete this property")
	}

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
