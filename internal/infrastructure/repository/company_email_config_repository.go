package repository

import (
	"context"
	"time"

	"github.com/myestatia/myestatia-go/internal/domain/entity"
	"gorm.io/gorm"
)

// CompanyEmailConfigRepository manages company email configurations
type CompanyEmailConfigRepository interface {
	Create(config *entity.CompanyEmailConfig) error
	Update(config *entity.CompanyEmailConfig) error
	FindByID(id string) (*entity.CompanyEmailConfig, error)
	FindByCompanyID(ctx context.Context, companyID string) (*entity.CompanyEmailConfig, error)
	FindAllEnabled(ctx context.Context) ([]*entity.CompanyEmailConfig, error)
	Delete(id string) error
	UpdateLastSync(ctx context.Context, id string, syncTime time.Time) error
}

type companyEmailConfigRepository struct {
	db *gorm.DB
}

// NewCompanyEmailConfigRepository creates a new repository
func NewCompanyEmailConfigRepository(db *gorm.DB) CompanyEmailConfigRepository {
	return &companyEmailConfigRepository{db: db}
}

func (r *companyEmailConfigRepository) Create(config *entity.CompanyEmailConfig) error {
	return r.db.Create(config).Error
}

func (r *companyEmailConfigRepository) Update(config *entity.CompanyEmailConfig) error {
	return r.db.Save(config).Error
}

func (r *companyEmailConfigRepository) FindByID(id string) (*entity.CompanyEmailConfig, error) {
	var config entity.CompanyEmailConfig
	err := r.db.First(&config, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &config, nil
}

func (r *companyEmailConfigRepository) FindByCompanyID(ctx context.Context, companyID string) (*entity.CompanyEmailConfig, error) {
	var config entity.CompanyEmailConfig
	err := r.db.WithContext(ctx).
		Where("company_id = ?", companyID).
		First(&config).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &config, nil
}

func (r *companyEmailConfigRepository) FindAllEnabled(ctx context.Context) ([]*entity.CompanyEmailConfig, error) {
	var configs []*entity.CompanyEmailConfig
	err := r.db.WithContext(ctx).
		Where("is_enabled = ?", true).
		Preload("Company").
		Find(&configs).Error

	if err != nil {
		return nil, err
	}
	return configs, nil
}

func (r *companyEmailConfigRepository) Delete(id string) error {
	return r.db.Unscoped().Delete(&entity.CompanyEmailConfig{}, "id = ?", id).Error
}

func (r *companyEmailConfigRepository) UpdateLastSync(ctx context.Context, id string, syncTime time.Time) error {
	return r.db.WithContext(ctx).
		Model(&entity.CompanyEmailConfig{}).
		Where("id = ?", id).
		Update("last_sync_at", syncTime).Error
}
