package repository

import (
	"context"
	"time"

	"github.com/myestatia/myestatia-go/internal/domain/entity"
	"gorm.io/gorm"
)

type PasswordResetRepository interface {
	Create(reset *entity.PasswordReset) error
	FindByToken(ctx context.Context, token string) (*entity.PasswordReset, error)
	MarkAsUsed(ctx context.Context, id string) error
	DeleteExpired(ctx context.Context) error
}

type passwordResetRepository struct {
	db *gorm.DB
}

func NewPasswordResetRepository(db *gorm.DB) PasswordResetRepository {
	return &passwordResetRepository{db: db}
}

func (r *passwordResetRepository) Create(reset *entity.PasswordReset) error {
	return r.db.Create(reset).Error
}

func (r *passwordResetRepository) FindByToken(ctx context.Context, token string) (*entity.PasswordReset, error) {
	var reset entity.PasswordReset
	err := r.db.WithContext(ctx).
		Preload("User").
		Where("token = ?", token).
		First(&reset).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return &reset, nil
}

func (r *passwordResetRepository) MarkAsUsed(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Model(&entity.PasswordReset{}).
		Where("id = ?", id).
		Update("used", true).Error
}

func (r *passwordResetRepository) DeleteExpired(ctx context.Context) error {
	return r.db.WithContext(ctx).
		Where("expires_at < ?", time.Now()).
		Delete(&entity.PasswordReset{}).Error
}
