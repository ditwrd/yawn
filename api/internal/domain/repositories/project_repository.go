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
// Uses Repository pattern with GORM ORM for project data persistence and
// retrieval.
// Supports PostgreSQL/SQLite with soft deletes and interface-based testing.
package repositories

import (
	"github.com/ditwrd/yawn/api/internal/domain/models"
	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

// ProjectRepository defines project data persistence operations.
//
// Supports CRUD operations with soft deletes, member management, and
// access control queries. Interface enables mocking for unit testing
// without database dependencies.
type ProjectRepository interface {
	// Create persists a new project to the database with auto-generated UUID.
	Create(project *models.Project) error

	// GetByID retrieves a project by UUID. Respects soft deletes.
	GetByID(id string) (*models.Project, error)

	// GetByIDWithMembers retrieves a project by UUID with its members preloaded.
	GetByIDWithMembers(id string) (*models.Project, error)

	// GetByUserID retrieves projects that a user has access to.
	GetByUserID(userID string, limit, offset int) ([]models.Project, error)

	// GetByOwnerID retrieves projects owned by a specific user.
	GetByOwnerID(ownerID string, limit, offset int) ([]models.Project, error)

	// List retrieves a paginated list of all projects. Respects soft deletes.
	List(limit, offset int) ([]models.Project, error)

	// Update updates an existing project record. Auto-updates UpdatedAt
	// timestamp.
	Update(project *models.Project) error

	// Delete soft deletes a project by UUID by setting DeletedAt timestamp.
	Delete(id string) error

	// AddMember adds a user to a project with a specific role.
	AddMember(projectID, userID string, role models.ProjectRole) error

	// RemoveMember removes a user from a project.
	RemoveMember(projectID, userID string) error

	// UpdateMemberRole updates a user's role in a project.
	UpdateMemberRole(projectID, userID string, role models.ProjectRole) error

	// GetMember retrieves a project member by project and user IDs.
	GetMember(projectID, userID string) (*models.ProjectUser, error)

	// ListMembers retrieves all members of a project.
	ListMembers(projectID string) ([]models.ProjectUser, error)

	// GetUserRole retrieves a user's role in a project.
	GetUserRole(projectID, userID string) (models.ProjectRole, error)

	// Search searches projects by name or description.
	Search(query string, limit, offset int) ([]models.Project, error)

	// Count returns the total count of projects.
	Count() (int64, error)

	// CountByUserID returns the count of projects accessible by a user.
	CountByUserID(userID string) (int64, error)
}

// projectRepository provides GORM-based implementation of ProjectRepository
// interface.
// Supports PostgreSQL/SQLite with soft deletes and connection pooling.
type projectRepository struct {
	// db provides GORM database connection for project operations
	db *gorm.DB
}

// NewProjectRepository creates a new ProjectRepository instance with provided
// GORM
// connection.
func NewProjectRepository(db *gorm.DB) ProjectRepository {
	return &projectRepository{db: db}
}

func (r *projectRepository) Create(project *models.Project) error {
	return r.db.Create(project).Error
}

func (r *projectRepository) GetByID(id string) (*models.Project, error) {
	var project models.Project

	err := r.db.Preload("Owner").Where("id = ?", id).First(&project).Error
	if err != nil {
		return nil, err
	}

	return &project, nil
}

func (r *projectRepository) GetByIDWithMembers(
	id string,
) (*models.Project, error) {
	var project models.Project

	err := r.db.Preload("Owner").
		Preload("Users.User").
		Where("id = ?", id).
		First(&project).
		Error
	if err != nil {
		return nil, err
	}

	return &project, nil
}

func (r *projectRepository) GetByUserID(
	userID string,
	limit, offset int,
) ([]models.Project, error) {
	var projects []models.Project

	// Get projects where user is owner or member
	err := r.db.Preload("Owner").
		Joins("LEFT JOIN project_users ON projects.id = project_users.project_id").
		Where("projects.owner_id = ? OR project_users.user_id = ?", userID, userID).
		Limit(limit).
		Offset(offset).
		Group("projects.id").
		Find(&projects).Error

	return projects, err
}

func (r *projectRepository) GetByOwnerID(
	ownerID string,
	limit, offset int,
) ([]models.Project, error) {
	var projects []models.Project

	err := r.db.Preload("Owner").
		Where("owner_id = ?", ownerID).
		Limit(limit).
		Offset(offset).
		Find(&projects).Error

	return projects, err
}

func (r *projectRepository) List(limit, offset int) ([]models.Project, error) {
	var projects []models.Project

	err := r.db.Preload("Owner").
		Limit(limit).
		Offset(offset).
		Find(&projects).Error

	return projects, err
}

func (r *projectRepository) Update(project *models.Project) error {
	return r.db.Save(project).Error
}

func (r *projectRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&models.Project{}).Error
}

func (r *projectRepository) AddMember(
	projectID, userID string,
	role models.ProjectRole,
) error {
	projectUUID, err := uuid.FromString(projectID)
	if err != nil {
		return err
	}

	userUUID, err := uuid.FromString(userID)
	if err != nil {
		return err
	}

	projectUser := models.ProjectUser{
		ProjectID: projectUUID,
		UserID:    userUUID,
		Role:      role,
	}

	return r.db.Create(&projectUser).Error
}

func (r *projectRepository) RemoveMember(projectID, userID string) error {
	return r.db.Where("project_id = ? AND user_id = ?", projectID, userID).
		Delete(&models.ProjectUser{}).Error
}

func (r *projectRepository) UpdateMemberRole(
	projectID, userID string,
	role models.ProjectRole,
) error {
	return r.db.Model(&models.ProjectUser{}).
		Where("project_id = ? AND user_id = ?", projectID, userID).
		Update("role", role).Error
}

func (r *projectRepository) GetMember(
	projectID, userID string,
) (*models.ProjectUser, error) {
	var member models.ProjectUser

	err := r.db.Preload("User").
		Where("project_id = ? AND user_id = ?", projectID, userID).
		First(&member).Error
	if err != nil {
		return nil, err
	}

	return &member, nil
}

func (r *projectRepository) ListMembers(
	projectID string,
) ([]models.ProjectUser, error) {
	var members []models.ProjectUser

	err := r.db.Preload("User").
		Where("project_id = ?", projectID).
		Find(&members).Error

	return members, err
}

func (r *projectRepository) GetUserRole(
	projectID, userID string,
) (models.ProjectRole, error) {
	var member models.ProjectUser

	err := r.db.Select("role").
		Where("project_id = ? AND user_id = ?", projectID, userID).
		First(&member).Error
	if err != nil {
		return "", err
	}

	return member.Role, nil
}

func (r *projectRepository) Search(
	query string,
	limit, offset int,
) ([]models.Project, error) {
	var projects []models.Project

	searchPattern := "%" + query + "%"

	err := r.db.Preload("Owner").
		Where("name ILIKE ? OR description ILIKE ?", searchPattern, searchPattern).
		Limit(limit).
		Offset(offset).
		Find(&projects).Error

	return projects, err
}

func (r *projectRepository) Count() (int64, error) {
	var count int64

	err := r.db.Model(&models.Project{}).Count(&count).Error

	return count, err
}

func (r *projectRepository) CountByUserID(userID string) (int64, error) {
	var count int64

	err := r.db.Model(&models.Project{}).
		Joins("LEFT JOIN project_users ON projects.id = project_users.project_id").
		Where("projects.owner_id = ? OR project_users.user_id = ?", userID, userID).
		Group("projects.id").
		Count(&count).Error

	return count, err
}
