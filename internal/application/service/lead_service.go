package service

import (
	"context"
	"errors"

	"bitbucket.org/statia/server/internal/domain/entity"
	"bitbucket.org/statia/server/internal/infrastructure/repository"
	"github.com/google/uuid"
)

type LeadService struct {
	Repo repository.LeadRepository
}

func NewLeadService(repo repository.LeadRepository) *LeadService {
	return &LeadService{Repo: repo}
}

// Create crea un nuevo lead (valida duplicados)
func (s *LeadService) Create(ctx context.Context, l *entity.Lead) (*entity.Lead, bool, error) {
	if l.Email == "" {
		return nil, false, errors.New("email is required")
	}

	existing, err := s.Repo.FindByEmail(ctx, l.Email)
	if err != nil {
		return nil, false, err
	}
	if existing != nil {
		return existing, false, nil
	}

	if l.ID == "" {
		l.ID = uuid.New().String()
	}

	if err := s.Repo.Create(l); err != nil {
		return nil, false, err
	}

	return l, true, nil
}

func (s *LeadService) FindByID(ctx context.Context, id string) (*entity.Lead, error) {
	return s.Repo.FindByID(id)
}

func (s *LeadService) FindAll(ctx context.Context) ([]entity.Lead, error) {
	return s.Repo.FindAll()
}

func (s *LeadService) Update(ctx context.Context, l *entity.Lead) error {
	if l.ID == "" {
		return errors.New("missing ID")
	}
	return s.Repo.Update(l)
}

func (s *LeadService) Delete(ctx context.Context, id string) error {
	return s.Repo.Delete(id)
}

func (s *LeadService) FindByCompanyId(ctx context.Context, id string) ([]entity.Lead, error) {
	return s.Repo.FindByCompanyId(ctx, id)
}

func (s *LeadService) FindByPropertyId(ctx context.Context, id string) ([]entity.Lead, error) {
	return s.Repo.FindByPropertyId(ctx, id)
}
