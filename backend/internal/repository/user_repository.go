package repository

import (
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/gigmatch/gigmatch/internal/model"
)

// UserRepository persists users.
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository builds a UserRepository.
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create inserts a user.
func (r *UserRepository) Create(u *model.User) error {
	if err := r.db.Create(u).Error; err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

// FindByUsername loads a user by username.
func (r *UserRepository) FindByUsername(username string) (*model.User, error) {
	var u model.User
	if err := r.db.Where("username = ?", username).First(&u).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("find user by username: %w", err)
	}
	return &u, nil
}

// ExistsByUsername reports whether a username is taken.
func (r *UserRepository) ExistsByUsername(username string) (bool, error) {
	var count int64
	if err := r.db.Model(&model.User{}).Where("username = ?", username).Count(&count).Error; err != nil {
		return false, fmt.Errorf("count users: %w", err)
	}
	return count > 0, nil
}

// FindByID loads a user by primary key.
func (r *UserRepository) FindByID(id uint) (*model.User, error) {
	var u model.User
	if err := r.db.First(&u, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("find user by id: %w", err)
	}
	return &u, nil
}

// Update persists user changes.
func (r *UserRepository) Update(u *model.User) error {
	if err := r.db.Save(u).Error; err != nil {
		return fmt.Errorf("update user: %w", err)
	}
	return nil
}
