package persistence

import (
	"github.com/unillm/unillm/internal/model"
	"gorm.io/gorm"
)

type UserRepo struct {
	db *gorm.DB
}

func NewUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) Create(user *model.User) error {
	return r.db.Create(user).Error
}

func (r *UserRepo) FindByEmail(email string) (*model.User, error) {
	var user model.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepo) FindByID(id int64) (*model.User, error) {
	var user model.User
	err := r.db.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepo) UpdateBalance(id int64, delta float64) error {
	return r.db.Model(&model.User{}).Where("id = ?", id).
		Update("balance", gorm.Expr("balance + ?", delta)).Error
}

func (r *UserRepo) UpdatePassword(id int64, passwordHash string) error {
	return r.db.Model(&model.User{}).Where("id = ?", id).
		Update("password_hash", passwordHash).Error
}

func (r *UserRepo) ListAll() ([]model.User, error) {
	var users []model.User
	err := r.db.Select("id, email, name, role, balance, created_at, updated_at").
		Order("id ASC").Find(&users).Error
	return users, err
}

func (r *UserRepo) GetBalance(id int64) (float64, error) {
	var user model.User
	if err := r.db.Select("balance").First(&user, id).Error; err != nil {
		return 0, err
	}
	return user.Balance, nil
}
