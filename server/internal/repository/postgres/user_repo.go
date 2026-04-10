package postgres

import (
	"context"
	"gorm.io/gorm"
	"vehivle/internal/domain/model"
)
type UserRepo struct {
	db *gorm.DB
}

func NewUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{db: db}
}
func (r *UserRepo) FindByID(ctx context.Context, id string) (*model.AdminUser, error) {
	var user model.AdminUser
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
func (r *UserRepo) FindByUsername(ctx context.Context, username string) (*model.AdminUser, error) {
	var user model.AdminUser
	if err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepo) Create(ctx context.Context, user *model.AdminUser) error {
	return r.db.WithContext(ctx).Create(user).Error
}