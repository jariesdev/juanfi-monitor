package repository

import (
	"strings"

	"github.com/jariesdev/vendoreport/internal/models"
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
// NOTE: passwords are stored as plain text to match the existing Python app.
func (r *UserRepository) CheckUser(username, password string) (*models.User, error) {
	var user models.User
	result := r.db.
		Where("username = ? AND password = ?", username, password).
		First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

// GetByUsername looks up a user by username, preloading their role and assigned vendos.
func (r *UserRepository) GetByUsername(username string) (*models.User, error) {
	var user models.User
	result := r.db.
		Preload("Role").
		Preload("Vendos").
		Where("username = ?", username).
		First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

// GetByID returns a single user by primary key, preloading role and vendos.
func (r *UserRepository) GetByID(id uint) (*models.User, error) {
	var user models.User
	if err := r.db.Preload("Role").Preload("Vendos").First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// List returns all users, optionally filtered by username substring.
func (r *UserRepository) List(q string, sortBy, sortDir string) ([]models.User, error) {
	var users []models.User

	query := r.db.Preload("Role")
	if q != "" {
		query = query.Where("username LIKE ?", "%"+q+"%")
	}

	allowed := map[string]bool{"username": true, "created_at": true, "is_active": true}
	col := "created_at"
	if allowed[sortBy] {
		col = sortBy
	}
	dir := "ASC"
	if strings.ToUpper(sortDir) == "DESC" {
		dir = "DESC"
	}
	query = query.Order(col + " " + dir)

	if err := query.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// Create inserts a new user record.
func (r *UserRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

// Update saves changes to an existing user record (all non-zero fields).
func (r *UserRepository) Update(user *models.User) error {
	return r.db.Save(user).Error
}

// Delete removes a user by ID.
func (r *UserRepository) Delete(id uint) error {
	return r.db.Delete(&models.User{}, id).Error
}

// UpdatePassword sets a new password for the given user ID.
func (r *UserRepository) UpdatePassword(userID uint, newPassword string) error {
	return r.db.Model(&models.User{}).Where("id = ?", userID).Update("password", newPassword).Error
}

// AssignVendos replaces the user's vendo assignments with the given vendo IDs.
func (r *UserRepository) AssignVendos(userID uint, vendoIDs []uint) error {
	user := models.User{ID: userID}
	var vendos []models.Vendo
	if len(vendoIDs) > 0 {
		if err := r.db.Find(&vendos, vendoIDs).Error; err != nil {
			return err
		}
	}
	return r.db.Model(&user).Association("Vendos").Replace(vendos)
}

// GetVendoIDs returns the IDs of all vendos assigned to the given user.
func (r *UserRepository) GetVendoIDs(userID uint) ([]uint, error) {
	var ids []uint
	err := r.db.Raw(
		"SELECT vendo_id FROM user_vendos WHERE user_id = ?", userID,
	).Scan(&ids).Error
	return ids, err
}
