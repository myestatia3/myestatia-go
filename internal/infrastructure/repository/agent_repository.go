package repository

import (
	"context"

	"github.com/myestatia/myestatia-go/internal/domain/entity"
	"gorm.io/gorm"
)

type AgentRepository interface {
	Create(agent *entity.Agent) error
	FindByID(id string) (*entity.Agent, error)
	FindAll() ([]entity.Agent, error)
	FindByCompanyID(companyID string) ([]entity.Agent, error)
	FindAdminByCompanyID(companyID string) (*entity.Agent, error)
	UpdatePartial(ctx context.Context, id string, fields map[string]interface{}) error
	Delete(id string) error
	FindByEmail(ctx context.Context, name string) (*entity.Agent, error)
}

type agentRepository struct {
	db *gorm.DB
}

func NewAgentRepository(db *gorm.DB) AgentRepository {
	return &agentRepository{db: db}
}

func (r *agentRepository) Create(agent *entity.Agent) error {
	return r.db.Create(agent).Error
}

func (r *agentRepository) FindByID(id string) (*entity.Agent, error) {
	var agent entity.Agent
	if err := r.db.First(&agent, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &agent, nil
}

func (r *agentRepository) FindAll() ([]entity.Agent, error) {
	var companies []entity.Agent
	if err := r.db.Find(&companies).Error; err != nil {
		return nil, err
	}
	return companies, nil
}

func (r *agentRepository) FindByCompanyID(companyID string) ([]entity.Agent, error) {
	var agents []entity.Agent
	if err := r.db.Where("company_id = ?", companyID).Find(&agents).Error; err != nil {
		return nil, err
	}
	return agents, nil
}

func (r *agentRepository) FindAdminByCompanyID(companyID string) (*entity.Agent, error) {
	var agent entity.Agent
	// Use Order and First to get the first admin (oldest by created_at) in case there are multiple
	if err := r.db.Where("company_id = ? AND role = ?", companyID, "admin").Order("created_at ASC").First(&agent).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // No admin found
		}
		return nil, err
	}
	return &agent, nil
}

func (r *agentRepository) UpdatePartial(ctx context.Context, id string, fields map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&entity.Agent{}).Where("id = ?", id).Updates(fields).Error
}

func (r *agentRepository) Delete(id string) error {
	return r.db.Unscoped().Delete(&entity.Agent{}, "id = ?", id).Error
}

func (r *agentRepository) FindByEmail(ctx context.Context, email string) (*entity.Agent, error) {
	var agent entity.Agent
	query := r.db.WithContext(ctx).Where("email = ?", email).First(&agent)

	if query.Error != nil {
		if query.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, query.Error
	}
	return &agent, nil
}
