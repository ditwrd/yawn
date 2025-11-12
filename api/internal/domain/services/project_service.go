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
	"github.com/gofrs/uuid"
)

// ProjectService provides project business operations.
type ProjectService interface {
	// Create creates a new project with the specified owner.
	Create(project *models.Project, ownerID string) error

	// GetByID retrieves a project by ID with access control check.
	GetByID(id, userID string) (*models.Project, error)

	// List retrieves projects accessible to the user with pagination.
	List(userID string, limit, offset int) ([]models.Project, error)

	// Search searches projects accessible to the user.
	Search(userID, query string, limit, offset int) ([]models.Project, error)

	// Update updates a project with access control check.
	Update(project *models.Project, userID string) error

	// Delete deletes a project with access control check.
	Delete(id, userID string) error

	// AddMember adds a member to a project with access control check.
	AddMember(
		projectID, userID, memberEmail, role string,
	) (*models.ProjectUser, error)

	// RemoveMember removes a member from a project with access control check.
	RemoveMember(projectID, userID, memberUserID string) error

	// UpdateMemberRole updates a member's role with access control check.
	UpdateMemberRole(
		projectID, userID, memberUserID, role string,
	) (*models.ProjectUser, error)

	// ListMembers lists all members of a project with access control check.
	ListMembers(projectID, userID string) ([]models.ProjectUser, error)

	// GetUserRole retrieves a user's role in a project.
	GetUserRole(projectID, userID string) (models.ProjectRole, error)

	// CheckAccess checks if a user has access to a project.
	CheckAccess(projectID, userID string, requiredRole models.ProjectRole) bool
}

// ProjectVisibility represents the visibility settings for a project.
type ProjectVisibility string

const (
	ProjectVisibilityPublic  ProjectVisibility = "public"
	ProjectVisibilityPrivate ProjectVisibility = "private"
)

// projectService implements ProjectService.
type projectService struct {
	projectRepo repositories.ProjectRepository
	userRepo    repositories.UserRepository
}

// NewProjectService creates a new project service.
func NewProjectService(
	projectRepo repositories.ProjectRepository,
	userRepo repositories.UserRepository,
) ProjectService {
	return &projectService{
		projectRepo: projectRepo,
		userRepo:    userRepo,
	}
}

func (s *projectService) Create(project *models.Project, ownerID string) error {
	if project == nil {
		return errors.New("project cannot be nil")
	}

	if ownerID == "" {
		return errors.New("owner ID cannot be empty")
	}

	// Validate project name
	project.Name = strings.TrimSpace(project.Name)
	if project.Name == "" {
		return errors.New("project name cannot be empty")
	}

	if len(project.Name) > 255 {
		return errors.New("project name too long (max 255 characters)")
	}

	// Trim and validate description
	if project.Description != "" {
		project.Description = strings.TrimSpace(project.Description)
		if len(project.Description) > 1000 {
			return errors.New("project description too long (max 1000 characters)")
		}
	}

	// Set default visibility to private
	if project.Visibility == "" {
		project.Visibility = string(ProjectVisibilityPrivate)
	} else {
		// Validate visibility
		validVisibilities := []ProjectVisibility{ProjectVisibilityPublic, ProjectVisibilityPrivate}

		visibility := ProjectVisibility(project.Visibility)
		if !slices.Contains(validVisibilities, visibility) {
			return fmt.Errorf("invalid project visibility: %s", project.Visibility)
		}
	}

	// Set owner ID
	ownerUUID, err := uuid.FromString(ownerID)
	if err != nil {
		return fmt.Errorf("invalid owner ID format: %w", err)
	}

	project.OwnerID = ownerUUID

	return s.projectRepo.Create(project)
}

func (s *projectService) GetByID(id, userID string) (*models.Project, error) {
	if id == "" {
		return nil, errors.New("project ID cannot be empty")
	}

	if userID == "" {
		return nil, errors.New("user ID cannot be empty")
	}

	project, err := s.projectRepo.GetByIDWithMembers(id)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve project: %w", err)
	}

	// Check access
	if !s.checkProjectAccess(project, userID) {
		return nil, fmt.Errorf("access denied to project %s", id)
	}

	return project, nil
}

func (s *projectService) List(
	userID string,
	limit, offset int,
) ([]models.Project, error) {
	if userID == "" {
		return nil, errors.New("user ID cannot be empty")
	}

	if limit <= 0 {
		limit = 20
	}

	if limit > 100 {
		limit = 100
	}

	if offset < 0 {
		offset = 0
	}

	projects, err := s.projectRepo.GetByUserID(userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}

	return projects, nil
}

func (s *projectService) Search(
	userID, query string,
	limit, offset int,
) ([]models.Project, error) {
	if userID == "" {
		return nil, errors.New("user ID cannot be empty")
	}

	if query == "" {
		return nil, errors.New("search query cannot be empty")
	}

	if len(query) < 2 {
		return nil, errors.New("search query too short (min 2 characters)")
	}

	if limit <= 0 {
		limit = 20
	}

	if limit > 100 {
		limit = 100
	}

	if offset < 0 {
		offset = 0
	}

	projects, err := s.projectRepo.Search(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search projects: %w", err)
	}

	// Filter projects accessible to the user
	var accessibleProjects []models.Project

	for _, project := range projects {
		if s.checkProjectAccess(&project, userID) {
			accessibleProjects = append(accessibleProjects, project)
		}
	}

	return accessibleProjects, nil
}

func (s *projectService) Update(project *models.Project, userID string) error {
	if project == nil {
		return errors.New("project cannot be nil")
	}

	if userID == "" {
		return errors.New("user ID cannot be empty")
	}

	if project.ID.String() == "" {
		return errors.New("project ID cannot be empty for update")
	}

	// Get existing project
	existingProject, err := s.projectRepo.GetByID(project.ID.String())
	if err != nil {
		return fmt.Errorf("project not found: %w", err)
	}

	// Check access (only owner or maintainer can update)
	if !s.CheckAccess(project.ID.String(), userID, models.ProjectRoleMaintainer) {
		return errors.New(
			"access denied: insufficient permissions to update project",
		)
	}

	// Validate and update fields if provided
	if project.Name != "" {
		project.Name = strings.TrimSpace(project.Name)
		if project.Name == "" {
			return errors.New("project name cannot be empty")
		}

		if len(project.Name) > 255 {
			return errors.New("project name too long (max 255 characters)")
		}

		existingProject.Name = project.Name
	}

	if project.Description != "" {
		project.Description = strings.TrimSpace(project.Description)
		if len(project.Description) > 1000 {
			return errors.New("project description too long (max 1000 characters)")
		}

		existingProject.Description = project.Description
	}

	if project.Visibility != "" {
		validVisibilities := []ProjectVisibility{
			ProjectVisibilityPublic,
			ProjectVisibilityPrivate,
		}

		visibility := ProjectVisibility(project.Visibility)
		if !slices.Contains(validVisibilities, visibility) {
			return fmt.Errorf("invalid project visibility: %s", project.Visibility)
		}

		existingProject.Visibility = string(visibility)
	}

	return s.projectRepo.Update(existingProject)
}

func (s *projectService) Delete(id, userID string) error {
	if id == "" {
		return errors.New("project ID cannot be empty")
	}

	if userID == "" {
		return errors.New("user ID cannot be empty")
	}

	// Check access (only owner can delete)
	if !s.CheckAccess(id, userID, models.ProjectRoleOwner) {
		return errors.New("access denied: only project owners can delete projects")
	}

	return s.projectRepo.Delete(id)
}

func (s *projectService) AddMember(
	projectID, userID, memberEmail, role string,
) (*models.ProjectUser, error) {
	if projectID == "" {
		return nil, errors.New("project ID cannot be empty")
	}

	if userID == "" {
		return nil, errors.New("user ID cannot be empty")
	}

	if memberEmail == "" {
		return nil, errors.New("member email cannot be empty")
	}

	// Validate role
	projectRole := models.ProjectRole(role)

	validRoles := []models.ProjectRole{
		models.ProjectRoleOwner,
		models.ProjectRoleMaintainer,
		models.ProjectRoleViewer,
	}
	if !slices.Contains(validRoles, projectRole) {
		return nil, fmt.Errorf("invalid project role: %s", role)
	}

	// Check access (only owner can add members)
	if !s.CheckAccess(projectID, userID, models.ProjectRoleOwner) {
		return nil, errors.New("access denied: only project owners can add members")
	}

	// Get the user to add
	memberUser, err := s.userRepo.GetByEmail(memberEmail)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Check if user is already a member
	existingMember, err := s.projectRepo.GetMember(
		projectID,
		memberUser.ID.String(),
	)
	if err == nil && existingMember != nil {
		return nil, errors.New("user is already a member of this project")
	}

	// Add the member
	err = s.projectRepo.AddMember(projectID, memberUser.ID.String(), projectRole)
	if err != nil {
		return nil, fmt.Errorf("failed to add member: %w", err)
	}

	// Return the created member
	return s.projectRepo.GetMember(projectID, memberUser.ID.String())
}

func (s *projectService) RemoveMember(
	projectID, userID, memberUserID string,
) error {
	if projectID == "" {
		return errors.New("project ID cannot be empty")
	}

	if userID == "" {
		return errors.New("user ID cannot be empty")
	}

	if memberUserID == "" {
		return errors.New("member user ID cannot be empty")
	}

	// Check access (only owner can remove members)
	if !s.CheckAccess(projectID, userID, models.ProjectRoleOwner) {
		return errors.New("access denied: only project owners can remove members")
	}

	// Check if member exists
	member, err := s.projectRepo.GetMember(projectID, memberUserID)
	if err != nil {
		return fmt.Errorf("member not found: %w", err)
	}

	// Prevent owners from removing themselves
	if member.UserID.String() == userID {
		return errors.New("cannot remove yourself from the project")
	}

	return s.projectRepo.RemoveMember(projectID, memberUserID)
}

func (s *projectService) UpdateMemberRole(
	projectID, userID, memberUserID, role string,
) (*models.ProjectUser, error) {
	if projectID == "" {
		return nil, errors.New("project ID cannot be empty")
	}

	if userID == "" {
		return nil, errors.New("user ID cannot be empty")
	}

	if memberUserID == "" {
		return nil, errors.New("member user ID cannot be empty")
	}

	// Validate role
	projectRole := models.ProjectRole(role)

	validRoles := []models.ProjectRole{
		models.ProjectRoleOwner,
		models.ProjectRoleMaintainer,
		models.ProjectRoleViewer,
	}
	if !slices.Contains(validRoles, projectRole) {
		return nil, fmt.Errorf("invalid project role: %s", role)
	}

	// Check access (only owner can update member roles)
	if !s.CheckAccess(projectID, userID, models.ProjectRoleOwner) {
		return nil, errors.New(
			"access denied: only project owners can update member roles",
		)
	}

	// Check if member exists
	member, err := s.projectRepo.GetMember(projectID, memberUserID)
	if err != nil {
		return nil, fmt.Errorf("member not found: %w", err)
	}

	// Update the role
	err = s.projectRepo.UpdateMemberRole(projectID, memberUserID, projectRole)
	if err != nil {
		return nil, fmt.Errorf("failed to update member role: %w", err)
	}

	// Return updated member
	member.Role = projectRole

	return member, nil
}

func (s *projectService) ListMembers(
	projectID, userID string,
) ([]models.ProjectUser, error) {
	if projectID == "" {
		return nil, errors.New("project ID cannot be empty")
	}

	if userID == "" {
		return nil, errors.New("user ID cannot be empty")
	}

	// Check access
	if !s.CheckAccess(projectID, userID, models.ProjectRoleViewer) {
		return nil, errors.New(
			"access denied: insufficient permissions to list members",
		)
	}

	return s.projectRepo.ListMembers(projectID)
}

func (s *projectService) GetUserRole(
	projectID, userID string,
) (models.ProjectRole, error) {
	if projectID == "" {
		return "", errors.New("project ID cannot be empty")
	}

	if userID == "" {
		return "", errors.New("user ID cannot be empty")
	}

	return s.projectRepo.GetUserRole(projectID, userID)
}

func (s *projectService) CheckAccess(
	projectID, userID string,
	requiredRole models.ProjectRole,
) bool {
	project, err := s.projectRepo.GetByID(projectID)
	if err != nil {
		return false
	}

	return s.checkProjectAccess(project, userID) &&
		s.hasMinimumRole(project, userID, requiredRole)
}

// checkProjectAccess checks if a user has any access to a project.
func (s *projectService) checkProjectAccess(
	project *models.Project,
	userID string,
) bool {
	// Owner has access
	if project.OwnerID.String() == userID {
		return true
	}

	// Check if user is a member
	role, err := s.projectRepo.GetUserRole(project.ID.String(), userID)
	if err != nil {
		return false
	}

	return role != ""
}

// hasMinimumRole checks if a user has at least the specified role in a project.
func (s *projectService) hasMinimumRole(
	project *models.Project,
	userID string,
	requiredRole models.ProjectRole,
) bool {
	// Owner has all permissions
	if project.OwnerID.String() == userID {
		return true
	}

	userRole, err := s.projectRepo.GetUserRole(project.ID.String(), userID)
	if err != nil {
		return false
	}

	// Role hierarchy: owner > maintainer > viewer
	switch requiredRole {
	case models.ProjectRoleViewer:
		return userRole == models.ProjectRoleViewer ||
			userRole == models.ProjectRoleMaintainer ||
			userRole == models.ProjectRoleOwner
	case models.ProjectRoleMaintainer:
		return userRole == models.ProjectRoleMaintainer ||
			userRole == models.ProjectRoleOwner
	case models.ProjectRoleOwner:
		return userRole == models.ProjectRoleOwner
	default:
		return false
	}
}
