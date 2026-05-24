package repository

import (
	"github.com/unillm/unillm/internal/model"
	"gorm.io/gorm"
)

type APIKeyRepo struct {
	db *gorm.DB
}

func NewAPIKeyRepo(db *gorm.DB) *APIKeyRepo {
	return &APIKeyRepo{db: db}
}

func (r *APIKeyRepo) Create(key *model.APIKey) error {
	return r.db.Create(key).Error
}

func (r *APIKeyRepo) FindByHash(hash string) (*model.APIKey, error) {
	var key model.APIKey
	err := r.db.Where("key_hash = ? AND is_active = true", hash).First(&key).Error
	if err != nil {
		return nil, err
	}
	return &key, nil
}

func (r *APIKeyRepo) ListByUserID(userID int64) ([]model.APIKey, error) {
	var keys []model.APIKey
	err := r.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&keys).Error
	return keys, err
}

func (r *APIKeyRepo) Deactivate(id, userID int64) error {
	return r.db.Model(&model.APIKey{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("is_active", false).Error
}

func (r *APIKeyRepo) UpdateLastUsed(id int64) error {
	return r.db.Model(&model.APIKey{}).Where("id = ?", id).
		Update("last_used", gorm.Expr("NOW()")).Error
}
