package repository

import (
	"context"
	"time"

	"github.com/myestatia/myestatia-go/internal/domain/entity"
	"gorm.io/gorm"
)

type InvitationRepository interface {
	Create(ctx context.Context, invitation *entity.Invitation) error
	FindByToken(ctx context.Context, token string) (*entity.Invitation, error)
	FindByEmail(ctx context.Context, email string) ([]*entity.Invitation, error)
	MarkAsUsed(ctx context.Context, token string) error
	Delete(ctx context.Context, token string) error
	DeleteExpired(ctx context.Context) error
}

type invitationRepository struct {
	db *gorm.DB
}

func NewInvitationRepository(db *gorm.DB) InvitationRepository {
	return &invitationRepository{db: db}
}

func (r *invitationRepository) Create(ctx context.Context, invitation *entity.Invitation) error {
	return r.db.WithContext(ctx).Create(invitation).Error
}

func (r *invitationRepository) FindByToken(ctx context.Context, token string) (*entity.Invitation, error) {
	var invitation entity.Invitation
	err := r.db.WithContext(ctx).
		Preload("Company").
		Where("token = ?", token).
		First(&invitation).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &invitation, nil
}

func (r *invitationRepository) FindByEmail(ctx context.Context, email string) ([]*entity.Invitation, error) {
	var invitations []*entity.Invitation
	err := r.db.WithContext(ctx).
		Where("email = ?", email).
		Order("created_at DESC").
		Find(&invitations).Error

	if err != nil {
		return nil, err
	}
	return invitations, nil
}

func (r *invitationRepository) MarkAsUsed(ctx context.Context, token string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&entity.Invitation{}).
		Where("token = ?", token).
		Updates(map[string]interface{}{
			"used":    true,
			"used_at": now,
		}).Error
}

func (r *invitationRepository) Delete(ctx context.Context, token string) error {
	return r.db.WithContext(ctx).
		Where("token = ?", token).
		Delete(&entity.Invitation{}).Error
}

func (r *invitationRepository) DeleteExpired(ctx context.Context) error {
	return r.db.WithContext(ctx).
		Where("expires_at < ? AND used = ?", time.Now(), false).
		Delete(&entity.Invitation{}).Error
}
