package repository

import (
	"context"
	"errors"

	"bitbucket.org/statia/server/internal/domain/entity"
	"gorm.io/gorm"
)

type PropertyRepository interface {
	Create(property *entity.Property) error
	FindByID(id string) (*entity.Property, error)
	FindAll() ([]entity.Property, error)
	Update(property *entity.Property) error
	Delete(id string) error
	FindByReference(ctx context.Context, ref string) (*entity.Property, error)
	FindAllByCompanyID(ctx context.Context, companyID string) ([]entity.Property, error)
	Search(ctx context.Context, filter entity.PropertyFilter) ([]entity.Property, error)
}

type propertyRepository struct {
	db *gorm.DB
}

func NewPropertyRepository(db *gorm.DB) PropertyRepository {
	return &propertyRepository{db: db}
}

func (r *propertyRepository) Create(property *entity.Property) error {
	return r.db.Create(property).Error
}

func (r *propertyRepository) FindByID(id string) (*entity.Property, error) {
	var prop entity.Property
	err := r.db.First(&prop, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Si no se encuentra la propiedad, devolvemos nil sin error
			return nil, nil
		}
		// Otros errores de DB se propagan
		return nil, err
	}
	return &prop, nil
}

func (r *propertyRepository) FindAll() ([]entity.Property, error) {
	var props []entity.Property
	if err := r.db.Find(&props).Error; err != nil {
		return nil, err
	}
	return props, nil
}

func (r *propertyRepository) FindByReference(ctx context.Context, ref string) (*entity.Property, error) {
	var p entity.Property
	query := r.db.WithContext(ctx).Where("reference = ?", ref).First(&p)
	if query.Error != nil {
		if query.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, query.Error
	}
	return &p, nil
}

func (r *propertyRepository) Update(property *entity.Property) error {
	return r.db.Save(property).Error
}

func (r *propertyRepository) Delete(id string) error {
	return r.db.Unscoped().Delete(&entity.Property{}, "id = ?", id).Error
}

func (r *propertyRepository) FindAllByCompanyID(ctx context.Context, companyID string) ([]entity.Property, error) {
	var properties []entity.Property
	if err := r.db.WithContext(ctx).
		Where("company_id = ?", companyID).
		Preload("Company"). //Carga company por detrás
		Find(&properties).Error; err != nil {
		return nil, err
	}
	return properties, nil
}

func (r *propertyRepository) Search(ctx context.Context, filter entity.PropertyFilter) ([]entity.Property, error) {
	var properties []entity.Property

	query := r.db.WithContext(ctx)

	// Habitaciones
	if filter.MinRooms != nil {
		query = query.Where("rooms >= ?", *filter.MinRooms)
	}
	if filter.MaxRooms != nil {
		query = query.Where("rooms <= ?", *filter.MaxRooms)
	}

	// Precio
	if filter.MinPrice != nil {
		query = query.Where("price >= ?", *filter.MinPrice)
	}
	if filter.MaxPrice != nil {
		query = query.Where("price <= ?", *filter.MaxPrice)
	}

	// Área (m2)
	if filter.MinAreaM2 != nil {
		query = query.Where("area_m2 >= ?", *filter.MinAreaM2)
	}
	if filter.MaxAreaM2 != nil {
		query = query.Where("area_m2 <= ?", *filter.MaxAreaM2)
	}

	if filter.Province != nil && *filter.Province != "" {
		// Partial Search: "Barc" will find "Barcelona"
		pattern := "%" + *filter.Province + "%"
		query = query.Where("province ILIKE ?", pattern)
	}

	if filter.Address != nil && *filter.Address != "" {
		pattern := "%" + *filter.Address + "%"
		query = query.Where("address ILIKE ?", pattern)
	}

	// Pagination

	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}

	// offset
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	if err := query.Find(&properties).Error; err != nil {
		return nil, err
	}

	return properties, nil
}
