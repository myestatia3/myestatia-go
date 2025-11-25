package repository

import (
	"context"

	"bitbucket.org/statia/server/internal/domain/entity"
	"gorm.io/gorm"
)

type CompanyRepository interface {
	Create(company *entity.Company) error
	FindByID(id string) (*entity.Company, error)
	FindAll() ([]entity.Company, error)
	UpdatePartial(ctx context.Context, id string, fields map[string]interface{}) error
	Delete(id string) error
	FindByName(ctx context.Context, name string) (*entity.Company, error)
}

type companyRepository struct {
	db *gorm.DB
}

func NewCompanyRepository(db *gorm.DB) CompanyRepository {
	return &companyRepository{db: db}
}

func (r *companyRepository) Create(company *entity.Company) error {
	return r.db.Create(company).Error
}

func (r *companyRepository) FindByID(id string) (*entity.Company, error) {
	var company entity.Company
	if err := r.db.First(&company, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &company, nil
}

func (r *companyRepository) FindAll() ([]entity.Company, error) {
	var companies []entity.Company
	if err := r.db.Find(&companies).Error; err != nil {
		return nil, err
	}
	return companies, nil
}

func (r *companyRepository) UpdatePartial(ctx context.Context, id string, fields map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&entity.Company{}).Where("id = ?", id).Updates(fields).Error
}

func (r *companyRepository) Delete(id string) error {
	return r.db.Unscoped().Delete(&entity.Company{}, "id = ?", id).Error
}

func (r *companyRepository) FindByName(ctx context.Context, name string) (*entity.Company, error) {
	var company entity.Company
	query := r.db.WithContext(ctx).Where("name = ?", name).First(&company)

	if query.Error != nil {
		if query.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, query.Error
	}
	return &company, nil
}
