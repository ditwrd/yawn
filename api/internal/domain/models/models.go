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

// Package models defines domain models for the YAWN application.
//
// Contains core entities: User, Project, Asset, Repository, Pipeline, and relationship models.
// Uses UUID v7 primary keys, GORM soft deletes, and supports PostgreSQL/SQLite.
package models

import (
	"time"

	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

// User represents a system user with authentication and authorization.
type User struct {
	ID           uuid.UUID      `gorm:"type:uuid;primary_key;default:uuid_generate_v7()" json:"id"`
	Email        string         `gorm:"uniqueIndex;not null"                             json:"email"`
	PasswordHash string         `gorm:"not null"                                         json:"-"`
	Role         UserRole       `gorm:"type:varchar(20);default:'user'"                  json:"role"`
	CreatedAt    time.Time      `gorm:"autoCreateTime"                                   json:"created_at"`
	UpdatedAt    time.Time      `gorm:"autoUpdateTime"                                   json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index"                                            json:"-"`
}

// UserRole represents the role of a user in the system.
type UserRole string

const (
	UserRoleAdmin UserRole = "admin"
	UserRoleUser  UserRole = "user"
)

// Project represents a project that groups assets, repositories, and users.
type Project struct {
	ID           uuid.UUID      `gorm:"type:uuid;primary_key;default:uuid_generate_v7()" json:"id"`
	Name         string         `gorm:"not null"                                         json:"name"`
	Description  string         `                                                        json:"description"`
	OwnerID      uuid.UUID      `gorm:"type:uuid;not null"                               json:"owner_id"`
	CreatedAt    time.Time      `gorm:"autoCreateTime"                                   json:"created_at"`
	UpdatedAt    time.Time      `gorm:"autoUpdateTime"                                   json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index"                                            json:"-"`
	Owner        User           `gorm:"foreignKey:OwnerID"                               json:"owner,omitempty"`
	Users        []ProjectUser  `gorm:"foreignKey:ProjectID"                             json:"users,omitempty"`
	Assets       []Asset        `gorm:"foreignKey:ProjectID"                             json:"assets,omitempty"`
	Repositories []Repository   `gorm:"foreignKey:ProjectID"                             json:"repositories,omitempty"`
}

// Asset represents a deployable asset within a project.
type Asset struct {
	ID           uuid.UUID      `gorm:"type:uuid;primary_key;default:uuid_generate_v7()" json:"id"`
	Name         string         `gorm:"not null"                                         json:"name"`
	Description  string         `                                                        json:"description"`
	Version      string         `gorm:"not null"                                         json:"version"`
	ProjectID    uuid.UUID      `gorm:"type:uuid;not null"                               json:"project_id"`
	RepositoryID *uuid.UUID     `gorm:"type:uuid"                                        json:"repository_id"`
	CreatedAt    time.Time      `gorm:"autoCreateTime"                                   json:"created_at"`
	UpdatedAt    time.Time      `gorm:"autoUpdateTime"                                   json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index"                                            json:"-"`

	Project    Project     `gorm:"foreignKey:ProjectID"       json:"project,omitempty"`
	Repository *Repository `gorm:"foreignKey:RepositoryID"    json:"repository,omitempty"`
	Pipelines  []Pipeline  `gorm:"many2many:asset_pipelines;" json:"pipelines,omitempty"`
}

// Repository represents a Git repository connected to a project.
type Repository struct {
	ID           uuid.UUID        `gorm:"type:uuid;primary_key;default:uuid_generate_v7()" json:"id"`
	URL          string           `gorm:"not null"                                         json:"url"`
	Branch       string           `gorm:"default:'main'"                                   json:"branch"`
	LatestCommit string           `                                                        json:"latest_commit"`
	SyncStatus   RepositoryStatus `gorm:"type:varchar(20);default:'pending'"               json:"sync_status"`
	ProjectID    uuid.UUID        `gorm:"type:uuid;not null"                               json:"project_id"`
	CreatedAt    time.Time        `gorm:"autoCreateTime"                                   json:"created_at"`
	UpdatedAt    time.Time        `gorm:"autoUpdateTime"                                   json:"updated_at"`
	DeletedAt    gorm.DeletedAt   `gorm:"index"                                            json:"-"`

	Project Project `gorm:"foreignKey:ProjectID"    json:"project,omitempty"`
	Assets  []Asset `gorm:"foreignKey:RepositoryID" json:"assets,omitempty"`
}

// RepositoryStatus represents the synchronization status of a repository.
type RepositoryStatus string

const (
	RepositoryStatusPending RepositoryStatus = "pending"
	RepositoryStatusSuccess RepositoryStatus = "success"
	RepositoryStatusError   RepositoryStatus = "error"
)

// Pipeline represents a deployment pipeline for processing assets.
type Pipeline struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:uuid_generate_v7()" json:"id"`
	Name        string         `gorm:"not null"                                         json:"name"`
	Description string         `                                                        json:"description"`
	ProjectID   uuid.UUID      `gorm:"type:uuid;not null"                               json:"project_id"`
	CreatedAt   time.Time      `gorm:"autoCreateTime"                                   json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime"                                   json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index"                                            json:"-"`

	Project Project `gorm:"foreignKey:ProjectID"       json:"project,omitempty"`
	Assets  []Asset `gorm:"many2many:asset_pipelines;" json:"assets,omitempty"`
}

// ProjectUser represents the many-to-many relationship between projects and users.
type ProjectUser struct {
	ID        uuid.UUID   `gorm:"type:uuid;primary_key;default:uuid_generate_v7()" json:"id"`
	ProjectID uuid.UUID   `gorm:"type:uuid;not null"                               json:"project_id"`
	UserID    uuid.UUID   `gorm:"type:uuid;not null"                               json:"user_id"`
	Role      ProjectRole `gorm:"type:varchar(20);not null"                        json:"role"`
	CreatedAt time.Time   `gorm:"autoCreateTime"                                   json:"created_at"`
	UpdatedAt time.Time   `gorm:"autoUpdateTime"                                   json:"updated_at"`

	Project Project `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
	User    User    `gorm:"foreignKey:UserID"    json:"user,omitempty"`
}

// ProjectRole represents the role of a user within a project.
type ProjectRole string

const (
	ProjectRoleOwner      ProjectRole = "owner"
	ProjectRoleMaintainer ProjectRole = "maintainer"
	ProjectRoleViewer     ProjectRole = "viewer"
)

// AssetPipeline represents the many-to-many relationship between assets and pipelines.
type AssetPipeline struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v7()" json:"id"`
	PipelineID uuid.UUID `gorm:"type:uuid;not null"                               json:"pipeline_id"`
	AssetID    uuid.UUID `gorm:"type:uuid;not null"                               json:"asset_id"`
	Order      int       `gorm:"not null"                                         json:"order"`
	CreatedAt  time.Time `gorm:"autoCreateTime"                                   json:"created_at"`

	Pipeline Pipeline `gorm:"foreignKey:PipelineID" json:"pipeline,omitempty"`
	Asset    Asset    `gorm:"foreignKey:AssetID"    json:"asset,omitempty"`
}
