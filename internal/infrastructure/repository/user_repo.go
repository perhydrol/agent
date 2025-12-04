package repository

import (
	"context"
	"errors"

	"github.com/perhydrol/insurance-agent-backend/pkg/domain"
	"gorm.io/gorm"
)

type userRepo struct {
	db *gorm.DB
}

func NewUerRepository(db *gorm.DB) domain.UserRepository {
	return &userRepo{
		db: db,
	}
}

func (r *userRepo) RegisterUser(ctx context.Context, user *domain.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}
func (r *userRepo) GetUserName(ctx context.Context, username string) (*domain.User, error) {
	var user domain.User
	err := r.db.WithContext(ctx).Where("username=?", username).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &user, err
}
