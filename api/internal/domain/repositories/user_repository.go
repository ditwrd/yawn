/*
Copyright © 2025 Aditya Wardianto <hi@ditwrd.dev>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/

// Package repositories provides data access layer implementations.
//
// Uses Repository pattern with GORM ORM for user data persistence and
// retrieval.
// Supports PostgreSQL/SQLite with soft deletes and interface-based testing.
package repositories

import (
	"github.com/ditwrd/yawn/api/internal/domain/models"
	"gorm.io/gorm"
)

// UserRepository defines user data persistence operations.
//
// Supports CRUD operations with soft deletes and GORM error handling.
// Interface enables mocking for unit testing without database dependencies.
type UserRepository interface {
	// Create persists a new user to the database with auto-generated UUID.
	Create(user *models.User) error

	// GetByID retrieves a user by UUID. Respects soft deletes.
	GetByID(id string) (*models.User, error)

	// GetByEmail retrieves a user by email address. Respects soft deletes.
	GetByEmail(email string) (*models.User, error)

	// Update updates an existing user record. Auto-updates UpdatedAt timestamp.
	Update(user *models.User) error

	// Delete soft deletes a user by UUID by setting DeletedAt timestamp.
	Delete(id string) error

	// List retrieves a paginated list of users. Respects soft deletes.
	List(limit, offset int) ([]models.User, error)
}

// userRepository provides GORM-based implementation of UserRepository
// interface.
// Supports PostgreSQL/SQLite with soft deletes and connection pooling.
type userRepository struct {
	// db provides GORM database connection for user operations
	db *gorm.DB
}

// NewUserRepository creates a new UserRepository instance with provided GORM
// connection.
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) GetByID(id string) (*models.User, error) {
	var user models.User
	err := r.db.Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Update(user *models.User) error {
	return r.db.Save(user).Error
}

func (r *userRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&models.User{}).Error
}

func (r *userRepository) List(limit, offset int) ([]models.User, error) {
	var users []models.User
	err := r.db.Limit(limit).Offset(offset).Find(&users).Error
	return users, err
}
