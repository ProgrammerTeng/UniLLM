package repository

import (
	"github.com/unillm/unillm/internal/model"
	"gorm.io/gorm"
)

type ProviderRepo struct {
	db *gorm.DB
}

func NewProviderRepo(db *gorm.DB) *ProviderRepo {
	return &ProviderRepo{db: db}
}

func (r *ProviderRepo) ListActive() ([]model.Provider, error) {
	var providers []model.Provider
	err := r.db.Where("is_active = true").Find(&providers).Error
	return providers, err
}

func (r *ProviderRepo) FindByID(id int64) (*model.Provider, error) {
	var p model.Provider
	err := r.db.First(&p, id).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *ProviderRepo) FindByName(name string) (*model.Provider, error) {
	var p model.Provider
	err := r.db.Where("name = ?", name).First(&p).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *ProviderRepo) ListActiveKeys(providerID int64) ([]model.ProviderKey, error) {
	var keys []model.ProviderKey
	err := r.db.Where("provider_id = ? AND is_active = true", providerID).Find(&keys).Error
	return keys, err
}

func (r *ProviderRepo) ListActiveModels() ([]model.ModelConfig, error) {
	var models []model.ModelConfig
	err := r.db.Where("is_active = true").Find(&models).Error
	return models, err
}

func (r *ProviderRepo) FindModelByPublicName(name string) (*model.ModelConfig, error) {
	var m model.ModelConfig
	err := r.db.Where("public_name = ? AND is_active = true", name).First(&m).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}
