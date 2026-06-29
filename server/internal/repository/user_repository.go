package repository

import (
	"github.com/jariesdev/vendoreport/internal/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// UserRepository is the concrete implementation of UserRepositoryInterface.
type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// CheckUser returns the user if the username and password match.
func (r *UserRepository) CheckUser(username, password string) (*models.User, error) {
	var user models.User
	result := r.db.Where("username = ?", username).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByUsername looks up a user by username, used during token validation.
func (r *UserRepository) GetByUsername(username string) (*models.User, error) {
	var user models.User
	result := r.db.Where("username = ?", username).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

// Create inserts a new user record.
func (r *UserRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

// UpdatePassword hashes newPassword and saves it for the given user ID.
func (r *UserRepository) UpdatePassword(userID uint, newPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return r.db.Model(&models.User{}).Where("id = ?", userID).Update("password", string(hash)).Error
}
