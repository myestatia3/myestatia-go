package service

import (
	"context"
	"errors"

	"bitbucket.org/statia/server/internal/domain/entity"
	"bitbucket.org/statia/server/internal/infrastructure/repository"
	"github.com/google/uuid"
)

type CompanyService struct {
	Repo repository.CompanyRepository
}

func NewCompanyService(repo repository.CompanyRepository) *CompanyService {
	return &CompanyService{Repo: repo}
}

// Create crea una nueva empresa
func (s *CompanyService) Create(ctx context.Context, c *entity.Company) (*entity.Company, bool, error) {
	if c.Name == "" {
		return nil, false, errors.New("company name is required")
	}

	existing, err := s.Repo.FindByName(ctx, c.Name)
	if err != nil {
		return nil, false, err
	}
	if existing != nil {
		return existing, false, nil // ya existe
	}
	c.ID = uuid.New().String()

	if err := s.Repo.Create(c); err != nil {
		return nil, false, err
	}

	return c, true, nil
}

func (s *CompanyService) FindByID(ctx context.Context, id string) (*entity.Company, error) {
	return s.Repo.FindByID(id)
}

func (s *CompanyService) FindAll(ctx context.Context) ([]entity.Company, error) {
	return s.Repo.FindAll()
}

// Update parcial
func (s *CompanyService) UpdatePartial(ctx context.Context, id string, fields map[string]interface{}) error {
	if id == "" {
		return errors.New("missing company ID")
	}
	return s.Repo.UpdatePartial(ctx, id, fields)
}

func (s *CompanyService) Delete(ctx context.Context, id string) error {
	return s.Repo.Delete(id)
}
