package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/myestatia/myestatia-go/internal/domain/entity"
	"github.com/myestatia/myestatia-go/internal/infrastructure/repository"
)

type AgentService struct {
	Repo               repository.AgentRepository
	PropertyRepository repository.PropertyRepository
}

func NewAgentService(repo repository.AgentRepository, propertyRepository repository.PropertyRepository) *AgentService {
	return &AgentService{Repo: repo, PropertyRepository: propertyRepository}
}

// Create crea una nueva agent
func (s *AgentService) Create(ctx context.Context, c *entity.Agent) (*entity.Agent, bool, error) {
	if c.Email == "" {
		return nil, false, errors.New("agent email is required")
	}

	existing, err := s.Repo.FindByEmail(ctx, c.Email)
	if err != nil {
		return nil, false, err
	}
	if existing != nil {
		return existing, false, nil // ya existe
	}

	c.ID = uuid.New().String()

	if err := s.AssociateProperties(c); err != nil {
		return nil, false, err
	}

	if err := s.Repo.Create(c); err != nil {
		return nil, false, err
	}

	return c, true, nil
}

func (s *AgentService) FindByID(ctx context.Context, id string) (*entity.Agent, error) {
	return s.Repo.FindByID(id)
}

func (s *AgentService) FindAll(ctx context.Context) ([]entity.Agent, error) {
	return s.Repo.FindAll()
}

// Update parcial
func (s *AgentService) UpdatePartial(ctx context.Context, id string, fields map[string]interface{}) error {
	if id == "" {
		return errors.New("missing agent ID")
	}
	return s.Repo.UpdatePartial(ctx, id, fields)
}

func (s *AgentService) GetByEmail(ctx context.Context, email string) (*entity.Agent, error) {
	return s.Repo.FindByEmail(ctx, email)
}

func (s *AgentService) Delete(ctx context.Context, id string) error {
	return s.Repo.Delete(id)
}

// AssociateProperties filtra las propiedades que existen en la base de datos
func (s *AgentService) AssociateProperties(agent *entity.Agent) error {
	if len(agent.Properties) == 0 {
		return nil
	}

	var existing []*entity.Property
	for _, p := range agent.Properties {
		prop, err := s.PropertyRepository.FindByID(p.ID)
		if err != nil {
			return err // errores de DB reales se siguen propagando
		}
		if prop != nil {
			existing = append(existing, prop) // solo agrega si existe
		}
		// si prop == nil, simplemente se ignora
	}
	agent.Properties = existing
	return nil
}
