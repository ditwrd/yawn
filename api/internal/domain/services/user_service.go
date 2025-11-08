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

// Package services provides business logic layer for the application.
package services

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/ditwrd/yawn/api/internal/domain/models"
	"github.com/ditwrd/yawn/api/internal/domain/repositories"
)

// UserService provides user business operations.
type UserService interface {
	Create(user *models.User) error
	GetByID(id string) (*models.User, error)
	GetByEmail(email string) (*models.User, error)
	Update(user *models.User) error
	Delete(id string) error
	List(limit, offset int) ([]models.User, error)
}

// userService implements UserService.
type userService struct {
	userRepo repositories.UserRepository
}

// NewUserService creates a new user service.
func NewUserService(userRepo repositories.UserRepository) UserService {
	return &userService{
		userRepo: userRepo,
	}
}

func (s *userService) Create(user *models.User) error {
	if user == nil {
		return errors.New("user cannot be nil")
	}

	user.Email = strings.ToLower(strings.TrimSpace(user.Email))

	if !strings.Contains(user.Email, "@") ||
		!strings.Contains(user.Email, ".") {
		return errors.New("invalid email format")
	}

	validRoles := []models.UserRole{models.UserRoleAdmin, models.UserRoleUser}
	isValidRole := slices.Contains(validRoles, user.Role)

	if !isValidRole {
		return fmt.Errorf("invalid user role: %s", user.Role)
	}

	existingUser, err := s.userRepo.GetByEmail(user.Email)
	if err == nil && existingUser != nil {
		return fmt.Errorf("user with email %s already exists", user.Email)
	}

	return s.userRepo.Create(user)
}

func (s *userService) GetByID(id string) (*models.User, error) {
	if id == "" {
		return nil, errors.New("user ID cannot be empty")
	}

	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve user: %w", err)
	}

	return user, nil
}

func (s *userService) GetByEmail(email string) (*models.User, error) {
	if email == "" {
		return nil, errors.New("email cannot be empty")
	}

	email = strings.ToLower(strings.TrimSpace(email))

	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve user by email: %w", err)
	}

	return user, nil
}

func (s *userService) Update(user *models.User) error {
	if user == nil {
		return errors.New("user cannot be nil")
	}

	if user.ID.String() == "" {
		return errors.New("user ID cannot be empty for update")
	}

	existingUser, err := s.userRepo.GetByID(user.ID.String())
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	if existingUser == nil {
		return fmt.Errorf("user with ID %s not found", user.ID.String())
	}

	if user.Email != "" {
		user.Email = strings.ToLower(strings.TrimSpace(user.Email))
		if !strings.Contains(user.Email, "@") ||
			!strings.Contains(user.Email, ".") {
			return errors.New("invalid email format")
		}
	}

	return s.userRepo.Update(user)
}

func (s *userService) Delete(id string) error {
	if id == "" {
		return errors.New("user ID cannot be empty")
	}

	existingUser, err := s.userRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	if existingUser == nil {
		return fmt.Errorf("user with ID %s not found", id)
	}

	return s.userRepo.Delete(id)
}

func (s *userService) List(limit, offset int) ([]models.User, error) {
	if limit <= 0 {
		limit = 20
	}

	if limit > 100 {
		limit = 100
	}

	if offset < 0 {
		offset = 0
	}

	users, err := s.userRepo.List(limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	return users, nil
}
