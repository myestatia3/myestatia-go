package repository

import (
	"context"

	"github.com/myestatia/myestatia-go/internal/domain/entity"
	"gorm.io/gorm"
)

type LeadRepository interface {
	Create(lead *entity.Lead) error
	FindByID(id string) (*entity.Lead, error)
	FindAll() ([]entity.Lead, error)
	Update(lead *entity.Lead) error
	Delete(id string) error
	FindByEmail(ctx context.Context, email string) (*entity.Lead, error)
	FindByCompanyId(ctx context.Context, companyId string) ([]entity.Lead, error)
	FindByPropertyId(ctx context.Context, propertyId string) ([]entity.Lead, error)
}

type leadRepository struct {
	db *gorm.DB
}

func NewLeadRepository(db *gorm.DB) LeadRepository {
	return &leadRepository{db: db}
}

func (r *leadRepository) Create(lead *entity.Lead) error {
	return r.db.Create(lead).Error
}

func (r *leadRepository) FindByID(id string) (*entity.Lead, error) {
	var lead entity.Lead
	if err := r.db.Preload("Property").First(&lead, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &lead, nil
}

func (r *leadRepository) FindAll() ([]entity.Lead, error) {
	var leads []entity.Lead
	if err := r.db.Order("created_at DESC").Find(&leads).Error; err != nil {
		return nil, err
	}

	for i := range leads {
		leads[i].SuggestedPropertiesCount = 0
		if leads[i].PropertyID != nil && *leads[i].PropertyID != "" && leads[i].Status != "discarded" {
			leads[i].SuggestedPropertiesCount += 1
		}
	}
	return leads, nil
}

func (r *leadRepository) Update(lead *entity.Lead) error {
	return r.db.Save(lead).Error
}

func (r *leadRepository) Delete(id string) error {
	return r.db.Unscoped().Delete(&entity.Lead{}, "id = ?", id).Error
}

func (r *leadRepository) FindByEmail(ctx context.Context, email string) (*entity.Lead, error) {
	var lead entity.Lead
	query := r.db.WithContext(ctx).
		Where("email = ?", email).
		Preload("Property").
		First(&lead)

	if query.Error != nil {
		if query.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, query.Error
	}

	return &lead, nil
}

func (r *leadRepository) FindByCompanyId(ctx context.Context, companyID string) ([]entity.Lead, error) {
	var leads []entity.Lead
	if err := r.db.WithContext(ctx).
		Where("company_id = ?", companyID).
		Preload("Company").       //Carga company por detrás
		Order("created_at DESC"). // Newest first
		Find(&leads).Error; err != nil {
		return nil, err
	}
	// Calculate SuggestedPropertiesCount
	for i := range leads {
		leads[i].SuggestedPropertiesCount = 0
		if leads[i].PropertyID != nil && *leads[i].PropertyID != "" && leads[i].Status != "discarded" {
			leads[i].SuggestedPropertiesCount += 1
		}
	}
	return leads, nil
}

func (r *leadRepository) FindByPropertyId(ctx context.Context, propertyID string) ([]entity.Lead, error) {
	var leads []entity.Lead
	if err := r.db.WithContext(ctx).
		Where("property_id = ?", propertyID).
		Preload("Property"). //Carga property por detrás
		Find(&leads).Error; err != nil {
		return nil, err
	}
	return leads, nil
}
