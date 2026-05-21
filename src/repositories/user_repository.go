package repositories

import (
	"errors"
	"project-manager/src/models"

	"gorm.io/gorm"
)

// handles database transactions for the User model
type UserRepository struct {
	db *gorm.DB
}

// instantiate a new repository with the global database connection
func NewUserRepository(database *gorm.DB) *UserRepository {
	return &UserRepository{db: database}
}

// insert a new user record into the database
func (r *UserRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

// search for an existing user record using their email address
func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // return nil with no error if user doesn't exist
		}
		return nil, err
	}
	return &user, nil
}
