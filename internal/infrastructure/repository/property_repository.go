package repository

import (
	"context"
	"errors"
	"strings"

	"github.com/myestatia/myestatia-go/internal/domain/entity"
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
	FindSubtypes(ctx context.Context, propertyType string) ([]entity.PropertySubtype, error)
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

	// Explicit Filters
	if filter.Status != nil && *filter.Status != "" && *filter.Status != "all" && *filter.Status != "todos" {
		query = query.Where("status = ?", *filter.Status)
	}

	if filter.Origin != nil && *filter.Origin != "" && *filter.Origin != "all" && *filter.Origin != "todos" {
		origin := *filter.Origin
		// Simple mapping if needed, e.g., "owned" -> "MANUAL"
		if strings.ToUpper(origin) == "OWNED" {
			origin = string(entity.OriginManual)
		} else if strings.ToUpper(origin) == "RESALES" {
			origin = "RESALES" // Or whatever constant is used for imports
		}
		query = query.Where("origin = ?", origin)
	}

	// Global Search Term
	if filter.SearchTerm != nil && *filter.SearchTerm != "" {
		pattern := "%" + *filter.SearchTerm + "%"
		query = query.Where(
			r.db.Where("title ILIKE ?", pattern).
				Or("reference ILIKE ?", pattern).
				Or("address ILIKE ?", pattern).
				Or("city ILIKE ?", pattern).
				Or("province ILIKE ?", pattern).
				Or("zone ILIKE ?", pattern),
		)
	}

	// Range Filters
	if filter.MinRooms != nil {
		query = query.Where("rooms >= ?", *filter.MinRooms)
	}
	if filter.MaxRooms != nil {
		query = query.Where("rooms <= ?", *filter.MaxRooms)
	}
	if filter.MinPrice != nil {
		query = query.Where("price >= ?", *filter.MinPrice)
	}
	if filter.MaxPrice != nil {
		query = query.Where("price <= ?", *filter.MaxPrice)
	}
	if filter.MinAreaM2 != nil {
		query = query.Where("area_m2 >= ?", *filter.MinAreaM2)
	}
	if filter.MaxAreaM2 != nil {
		query = query.Where("area_m2 <= ?", *filter.MaxAreaM2)
	}

	// Legacy filters (keep for compatibility if needed, though global search overrides if both present)
	if filter.Province != nil && *filter.Province != "" {
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
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	if err := query.Find(&properties).Error; err != nil {
		return nil, err
	}

	return properties, nil
}

func (r *propertyRepository) FindSubtypes(ctx context.Context, propertyType string) ([]entity.PropertySubtype, error) {
	var subtypes []entity.PropertySubtype
	query := r.db.WithContext(ctx).Where("active = ?", true)

	if propertyType != "" {
		query = query.Where("type = ?", propertyType)
	}

	if err := query.Find(&subtypes).Error; err != nil {
		return nil, err
	}
	return subtypes, nil
}
