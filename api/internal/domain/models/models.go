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

// Package models defines the domain models for the YAWN application.
//
// This package contains the core domain entities that represent the business objects
// in the YAWN (Yet Another Workflow Navigator) system. These models are designed
// following Domain-Driven Design (DDD) principles and include:
//
// - User: Represents system users with authentication and authorization
// - Project: Represents projects that contain assets and repositories
// - Asset: Represents deployable assets within projects
// - Repository: Represents Git repositories connected to projects
// - Pipeline: Represents deployment pipelines for assets
// - ProjectUser: Represents many-to-many relationship between projects and users
// - AssetPipeline: Represents many-to-many relationship between assets and pipelines
//
// All models use UUID v7 as primary keys for better performance and chronological ordering.
// The models support soft deletes through GORM's DeletedAt field and include proper
// database relationships and constraints.
//
// Database schema:
// These models are designed to work with GORM ORM and support both PostgreSQL and SQLite
// databases. All relationships are properly defined with foreign key constraints.
//
// Example usage:
//
//	user := &models.User{
//		Email: "user@example.com",
//		PasswordHash: "hashed_password",
//		Role: models.UserRoleUser,
//	}
//
//	project := &models.Project{
//		Name: "My Project",
//		Description: "A sample project",
//		OwnerID: user.ID,
//	}
package models

import (
	"time"

	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

// User represents a user in the system.
//
// The User model stores authentication and authorization information for system users.
// It supports role-based access control and uses UUID v7 for the primary key.
//
// Fields:
//   - ID: Unique identifier using UUID v7 for chronological ordering
//   - Email: User's email address, must be unique
//   - PasswordHash: Hashed password (never exposed in JSON)
//   - Role: User role for authorization (admin or user)
//   - CreatedAt: Timestamp when the user was created
//   - UpdatedAt: Timestamp when the user was last updated
//   - DeletedAt: Soft delete timestamp (nullable)
//
// Database constraints:
//   - Email has a unique index
//   - Email and PasswordHash cannot be null
//   - Default role is "user"
//   - Soft delete supported through DeletedAt field
//
// JSON serialization:
//   - PasswordHash is excluded from JSON output
//   - DeletedAt is excluded from JSON output
//
// Example usage:
//
//	user := &models.User{
//		Email: "admin@example.com",
//		PasswordHash: "$2a$10$...", // bcrypt hash
//		Role: models.UserRoleAdmin,
//	}
type User struct {
	// ID is the unique identifier for the user using UUID v7
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v7()" json:"id"`
	// Email is the user's email address, must be unique
	Email        string    `gorm:"uniqueIndex;not null" json:"email"`
	// PasswordHash contains the bcrypt hash of the user's password, never exposed in JSON
	PasswordHash string    `gorm:"not null" json:"-"`
	// Role defines the user's authorization level (admin or user)
	Role         UserRole  `gorm:"type:varchar(20);default:'user'" json:"role"`
	// CreatedAt is the timestamp when the user record was created
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	// UpdatedAt is the timestamp when the user record was last updated
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	// DeletedAt is used for soft deletes, contains the deletion timestamp
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// UserRole represents the role of a user in the system.
//
// UserRole defines the authorization level for users and determines what
// actions they can perform within the application.
//
// Available roles:
//   - UserRoleAdmin: Administrative users with full system access
//   - UserRoleUser: Regular users with limited access
//
// Default role:
// New users are assigned the UserRoleUser role by default.
type UserRole string

const (
	// UserRoleAdmin represents an administrator with full system access
	UserRoleAdmin UserRole = "admin"
	// UserRoleUser represents a regular user with standard access
	UserRoleUser  UserRole = "user"
)

// Project represents a project in the system.
//
// The Project model is the central entity that groups together assets, repositories,
// and users. Each project is owned by a single user and can have multiple users
// with different permission levels.
//
// Fields:
//   - ID: Unique identifier using UUID v7
//   - Name: Project name, cannot be null
//   - Description: Optional project description
//   - OwnerID: Foreign key to the user who owns this project
//   - CreatedAt: Timestamp when the project was created
//   - UpdatedAt: Timestamp when the project was last updated
//   - DeletedAt: Soft delete timestamp
//
// Relationships:
//   - Owner: BelongsTo relationship with User model
//   - Users: HasMany relationship through ProjectUser (many-to-many)
//   - Assets: HasMany relationship with Asset model
//   - Repositories: HasMany relationship with Repository model
//
// Database constraints:
//   - Name cannot be null
//   - OwnerID cannot be null and has foreign key constraint
//   - Soft delete supported through DeletedAt field
//
// Example usage:
//
//	project := &models.Project{
//		Name: "My Web App",
//		Description: "A sample web application project",
//		OwnerID: ownerUser.ID,
//	}
type Project struct {
	// ID is the unique identifier for the project using UUID v7
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v7()" json:"id"`
	// Name is the project name, cannot be null
	Name        string    `gorm:"not null" json:"name"`
	// Description is an optional description of the project
	Description string    `json:"description"`
	// OwnerID is the foreign key to the user who owns this project
	OwnerID     uuid.UUID `gorm:"type:uuid;not null" json:"owner_id"`
	// CreatedAt is the timestamp when the project record was created
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	// UpdatedAt is the timestamp when the project record was last updated
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	// DeletedAt is used for soft deletes, contains the deletion timestamp
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	// Owner is the user who owns this project (BelongsTo relationship)
	Owner     User            `gorm:"foreignKey:OwnerID" json:"owner,omitempty"`
	// Users are the users associated with this project (HasMany through ProjectUser)
	Users     []ProjectUser   `gorm:"foreignKey:ProjectID" json:"users,omitempty"`
	// Assets are the assets belonging to this project (HasMany relationship)
	Assets    []Asset         `gorm:"foreignKey:ProjectID" json:"assets,omitempty"`
	// Repositories are the Git repositories connected to this project (HasMany relationship)
	Repositories []Repository `gorm:"foreignKey:ProjectID" json:"repositories,omitempty"`
}

// Asset represents a deployable asset in the system.
//
// The Asset model represents individual deployable components that can be
// associated with projects and pipelines. Assets can be linked to repositories
// and can be processed through multiple deployment pipelines.
//
// Fields:
//   - ID: Unique identifier using UUID v7
//   - Name: Asset name, cannot be null
//   - Description: Optional asset description
//   - Version: Asset version, cannot be null
//   - ProjectID: Foreign key to the project this asset belongs to
//   - RepositoryID: Optional foreign key to the source repository
//   - CreatedAt: Timestamp when the asset was created
//   - UpdatedAt: Timestamp when the asset was last updated
//   - DeletedAt: Soft delete timestamp
//
// Relationships:
//   - Project: BelongsTo relationship with Project model
//   - Repository: Optional BelongsTo relationship with Repository model
//   - Pipelines: Many-to-many relationship through AssetPipeline
//
// Database constraints:
//   - Name and Version cannot be null
//   - ProjectID cannot be null and has foreign key constraint
//   - RepositoryID is nullable and optional
//   - Soft delete supported through DeletedAt field
//
// Example usage:
//
//	asset := &models.Asset{
//		Name: "web-app",
//		Version: "1.0.0",
//		ProjectID: project.ID,
//		RepositoryID: &repo.ID,
//	}
type Asset struct {
	// ID is the unique identifier for the asset using UUID v7
	ID           uuid.UUID  `gorm:"type:uuid;primary_key;default:uuid_generate_v7()" json:"id"`
	// Name is the asset name, cannot be null
	Name         string     `gorm:"not null" json:"name"`
	// Description is an optional description of the asset
	Description  string     `json:"description"`
	// Version is the version of the asset, cannot be null
	Version      string     `gorm:"not null" json:"version"`
	// ProjectID is the foreign key to the project this asset belongs to
	ProjectID    uuid.UUID  `gorm:"type:uuid;not null" json:"project_id"`
	// RepositoryID is the optional foreign key to the source repository
	RepositoryID *uuid.UUID `gorm:"type:uuid" json:"repository_id"`
	// CreatedAt is the timestamp when the asset record was created
	CreatedAt    time.Time  `gorm:"autoCreateTime" json:"created_at"`
	// UpdatedAt is the timestamp when the asset record was last updated
	UpdatedAt    time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
	// DeletedAt is used for soft deletes, contains the deletion timestamp
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	// Project is the project this asset belongs to (BelongsTo relationship)
	Project     Project    `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
	// Repository is the optional source repository for this asset (BelongsTo relationship)
	Repository  *Repository `gorm:"foreignKey:RepositoryID" json:"repository,omitempty"`
	// Pipelines are the deployment pipelines that can process this asset (Many-to-many)
	Pipelines   []Pipeline  `gorm:"many2many:asset_pipelines;" json:"pipelines,omitempty"`
}

// Repository represents a Git repository connected to a project.
//
// The Repository model stores information about Git repositories that are
// connected to projects. It tracks synchronization status and metadata
// about the repository state.
//
// Fields:
//   - ID: Unique identifier using UUID v7
//   - URL: Git repository URL, cannot be null
//   - Branch: Git branch to track (default: "main")
//   - LatestCommit: Hash of the latest synchronized commit
//   - SyncStatus: Current synchronization status
//   - ProjectID: Foreign key to the associated project
//   - CreatedAt: Timestamp when the repository was added
//   - UpdatedAt: Timestamp when the repository was last updated
//   - DeletedAt: Soft delete timestamp
//
// Relationships:
//   - Project: BelongsTo relationship with Project model
//   - Assets: HasMany relationship with Asset model
//
// Database constraints:
//   - URL cannot be null
//   - ProjectID cannot be null and has foreign key constraint
//   - Default branch is "main"
//   - Default sync status is "pending"
//   - Soft delete supported through DeletedAt field
//
// Example usage:
//
//	repo := &models.Repository{
//		URL: "https://github.com/user/repo.git",
//		Branch: "main",
//		ProjectID: project.ID,
//	}
type Repository struct {
	// ID is the unique identifier for the repository using UUID v7
	ID           uuid.UUID       `gorm:"type:uuid;primary_key;default:uuid_generate_v7()" json:"id"`
	// URL is the Git repository URL, cannot be null
	URL          string          `gorm:"not null" json:"url"`
	// Branch is the Git branch to track (default: "main")
	Branch       string          `gorm:"default:'main'" json:"branch"`
	// LatestCommit is the hash of the latest synchronized commit
	LatestCommit string          `json:"latest_commit"`
	// SyncStatus is the current synchronization status of the repository
	SyncStatus   RepositoryStatus `gorm:"type:varchar(20);default:'pending'" json:"sync_status"`
	// ProjectID is the foreign key to the associated project
	ProjectID    uuid.UUID       `gorm:"type:uuid;not null" json:"project_id"`
	// CreatedAt is the timestamp when the repository record was created
	CreatedAt    time.Time       `gorm:"autoCreateTime" json:"created_at"`
	// UpdatedAt is the timestamp when the repository record was last updated
	UpdatedAt    time.Time       `gorm:"autoUpdateTime" json:"updated_at"`
	// DeletedAt is used for soft deletes, contains the deletion timestamp
	DeletedAt    gorm.DeletedAt  `gorm:"index" json:"-"`

	// Relationships
	// Project is the project this repository belongs to (BelongsTo relationship)
	Project Project `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
	// Assets are the assets that are linked to this repository (HasMany relationship)
	Assets  []Asset `gorm:"foreignKey:RepositoryID" json:"assets,omitempty"`
}

// RepositoryStatus represents the synchronization status of a repository.
//
// This enum defines the possible states of repository synchronization
// with the remote Git repository.
//
// Available statuses:
//   - RepositoryStatusPending: Repository is pending synchronization
//   - RepositoryStatusSuccess: Repository synchronization was successful
//   - RepositoryStatusError: Repository synchronization failed
//
// Default status:
// New repositories start with RepositoryStatusPending status.
type RepositoryStatus string

const (
	// RepositoryStatusPending indicates the repository is waiting to be synchronized
	RepositoryStatusPending RepositoryStatus = "pending"
	// RepositoryStatusSuccess indicates the repository was successfully synchronized
	RepositoryStatusSuccess RepositoryStatus = "success"
	// RepositoryStatusError indicates the repository synchronization failed
	RepositoryStatusError   RepositoryStatus = "error"
)

// Pipeline represents a deployment pipeline for processing assets.
//
// The Pipeline model defines deployment pipelines that can process assets
// within a project. Pipelines can be associated with multiple assets
// through a many-to-many relationship.
//
// Fields:
//   - ID: Unique identifier using UUID v7
//   - Name: Pipeline name, cannot be null
//   - Description: Optional pipeline description
//   - ProjectID: Foreign key to the associated project
//   - CreatedAt: Timestamp when the pipeline was created
//   - UpdatedAt: Timestamp when the pipeline was last updated
//   - DeletedAt: Soft delete timestamp
//
// Relationships:
//   - Project: BelongsTo relationship with Project model
//   - Assets: Many-to-many relationship through AssetPipeline
//
// Database constraints:
//   - Name cannot be null
//   - ProjectID cannot be null and has foreign key constraint
//   - Soft delete supported through DeletedAt field
//
// Example usage:
//
//	pipeline := &models.Pipeline{
//		Name: "deploy-to-production",
//		Description: "Production deployment pipeline",
//		ProjectID: project.ID,
//	}
type Pipeline struct {
	// ID is the unique identifier for the pipeline using UUID v7
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v7()" json:"id"`
	// Name is the pipeline name, cannot be null
	Name        string    `gorm:"not null" json:"name"`
	// Description is an optional description of the pipeline
	Description string    `json:"description"`
	// ProjectID is the foreign key to the associated project
	ProjectID   uuid.UUID `gorm:"type:uuid;not null" json:"project_id"`
	// CreatedAt is the timestamp when the pipeline record was created
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	// UpdatedAt is the timestamp when the pipeline record was last updated
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	// DeletedAt is used for soft deletes, contains the deletion timestamp
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	// Project is the project this pipeline belongs to (BelongsTo relationship)
	Project Project `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
	// Assets are the assets that can be processed by this pipeline (Many-to-many)
	Assets  []Asset `gorm:"many2many:asset_pipelines;" json:"assets,omitempty"`
}

// ProjectUser represents the many-to-many relationship between projects and users.
//
// This join table manages user permissions and access levels for projects.
// Each user can have different roles in different projects.
//
// Fields:
//   - ID: Unique identifier using UUID v7
//   - ProjectID: Foreign key to the project
//   - UserID: Foreign key to the user
//   - Role: User's role within this project
//   - CreatedAt: Timestamp when the relationship was created
//   - UpdatedAt: Timestamp when the relationship was last updated
//
// Relationships:
//   - Project: BelongsTo relationship with Project model
//   - User: BelongsTo relationship with User model
//
// Database constraints:
//   - ProjectID and UserID cannot be null
//   - Role cannot be null
//   - Composite unique index on (ProjectID, UserID) to prevent duplicates
//
// Example usage:
//
//	projectUser := &models.ProjectUser{
//		ProjectID: project.ID,
//		UserID: user.ID,
//		Role: models.ProjectRoleMaintainer,
//	}
type ProjectUser struct {
	// ID is the unique identifier for the project-user relationship using UUID v7
	ID        uuid.UUID    `gorm:"type:uuid;primary_key;default:uuid_generate_v7()" json:"id"`
	// ProjectID is the foreign key to the project
	ProjectID uuid.UUID    `gorm:"type:uuid;not null" json:"project_id"`
	// UserID is the foreign key to the user
	UserID    uuid.UUID    `gorm:"type:uuid;not null" json:"user_id"`
	// Role is the user's role within this project
	Role      ProjectRole  `gorm:"type:varchar(20);not null" json:"role"`
	// CreatedAt is the timestamp when the relationship was created
	CreatedAt time.Time    `gorm:"autoCreateTime" json:"created_at"`
	// UpdatedAt is the timestamp when the relationship was last updated
	UpdatedAt time.Time    `gorm:"autoUpdateTime" json:"updated_at"`

	// Relationships
	// Project is the project in this relationship (BelongsTo relationship)
	Project Project `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
	// User is the user in this relationship (BelongsTo relationship)
	User    User    `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// ProjectRole represents the role of a user within a project.
//
// This enum defines the permission levels that users can have within projects,
// determining what actions they can perform on project resources.
//
// Available roles:
//   - ProjectRoleOwner: Full control over the project, can manage all settings
//   - ProjectRoleMaintainer: Can modify project content and manage resources
//   - ProjectRoleViewer: Read-only access to project resources
//
// Permission hierarchy:
// Owner > Maintainer > Viewer
type ProjectRole string

const (
	// ProjectRoleOwner indicates full ownership and control over the project
	ProjectRoleOwner     ProjectRole = "owner"
	// ProjectRoleMaintainer indicates the user can maintain and modify project resources
	ProjectRoleMaintainer ProjectRole = "maintainer"
	// ProjectRoleViewer indicates read-only access to project resources
	ProjectRoleViewer    ProjectRole = "viewer"
)

// AssetPipeline represents the many-to-many relationship between assets and pipelines.
//
// This join table manages which pipelines can process which assets and in what order.
// It allows assets to be processed through multiple pipelines in a specific sequence.
//
// Fields:
//   - ID: Unique identifier using UUID v7
//   - PipelineID: Foreign key to the pipeline
//   - AssetID: Foreign key to the asset
//   - Order: Execution order for the pipeline (1, 2, 3, etc.)
//   - CreatedAt: Timestamp when the relationship was created
//
// Relationships:
//   - Pipeline: BelongsTo relationship with Pipeline model
//   - Asset: BelongsTo relationship with Asset model
//
// Database constraints:
//   - PipelineID and AssetID cannot be null
//   - Order cannot be null and should be positive
//   - Composite unique index on (PipelineID, AssetID) to prevent duplicates
//
// Execution order:
// The Order field determines the sequence in which pipelines are executed
// for a given asset. Lower numbers are executed first.
//
// Example usage:
//
//	assetPipeline := &models.AssetPipeline{
//		PipelineID: pipeline.ID,
//		AssetID: asset.ID,
//		Order: 1, // First pipeline to execute
//	}
type AssetPipeline struct {
	// ID is the unique identifier for the asset-pipeline relationship using UUID v7
	ID         uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v7()" json:"id"`
	// PipelineID is the foreign key to the pipeline
	PipelineID uuid.UUID `gorm:"type:uuid;not null" json:"pipeline_id"`
	// AssetID is the foreign key to the asset
	AssetID    uuid.UUID `gorm:"type:uuid;not null" json:"asset_id"`
	// Order is the execution order for this pipeline (starts from 1)
	Order      int       `gorm:"not null" json:"order"`
	// CreatedAt is the timestamp when the relationship was created
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`

	// Relationships
	// Pipeline is the pipeline in this relationship (BelongsTo relationship)
	Pipeline Pipeline `gorm:"foreignKey:PipelineID" json:"pipeline,omitempty"`
	// Asset is the asset in this relationship (BelongsTo relationship)
	Asset    Asset    `gorm:"foreignKey:AssetID" json:"asset,omitempty"`
}