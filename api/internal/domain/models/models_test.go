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

package models_test

import (
	"testing"

	"github.com/ditwrd/yawn/api/internal/domain/models"
	"github.com/ditwrd/yawn/api/internal/infrastructure/database"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: false,
	})
	require.NoError(t, err)

	// Run migrations
	err = database.Migrate(db)
	require.NoError(t, err)

	return db
}

// TestUserModelValidation tests User model validations and constraints
func TestUserModelValidation(t *testing.T) {
	db := setupTestDB(t)

	t.Run("create user with valid data", func(t *testing.T) {
		userID := uuid.Must(uuid.NewV7())
		user := &models.User{
			ID:           userID,
			Email:        "test@example.com",
			PasswordHash: "hashed_password",
			Role:         models.UserRoleUser,
		}

		err := db.Create(user).Error
		assert.NoError(t, err)
		assert.Equal(t, userID, user.ID)
		assert.Equal(t, "test@example.com", user.Email)
		assert.Equal(t, models.UserRoleUser, user.Role)
	})

	t.Run("enforce unique email constraint", func(t *testing.T) {
		user1 := &models.User{
			ID:           uuid.Must(uuid.NewV7()),
			Email:        "duplicate@example.com",
			PasswordHash: "hashed_password",
			Role:         models.UserRoleUser,
		}
		err := db.Create(user1).Error
		require.NoError(t, err)

		user2 := &models.User{
			ID:           uuid.Must(uuid.NewV7()),
			Email:        "duplicate@example.com", // Same email
			PasswordHash: "hashed_password2",
			Role:         models.UserRoleAdmin,
		}
		err = db.Create(user2).Error
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "UNIQUE")
	})
}

// TestProjectModelValidation tests Project model validations and constraints
func TestProjectModelValidation(t *testing.T) {
	db := setupTestDB(t)

	// Create a test user first
	user := &models.User{
		ID:           uuid.Must(uuid.NewV7()),
		Email:        "owner@example.com",
		PasswordHash: "hashed_password",
		Role:         models.UserRoleUser,
	}
	err := db.Create(user).Error
	require.NoError(t, err)

	t.Run("create project with valid data", func(t *testing.T) {
		project := &models.Project{
			ID:          uuid.Must(uuid.NewV7()),
			Name:        "Test Project",
			Description: "A test project",
			OwnerID:     user.ID,
		}

		err := db.Create(project).Error
		assert.NoError(t, err)
		assert.Equal(t, "Test Project", project.Name)
		assert.Equal(t, user.ID, project.OwnerID)
	})
}

// TestAssetModelValidation tests Asset model validations and constraints
func TestAssetModelValidation(t *testing.T) {
	db := setupTestDB(t)

	// Create test user and project
	user := &models.User{
		ID:           uuid.Must(uuid.NewV7()),
		Email:        "owner@example.com",
		PasswordHash: "hashed_password",
		Role:         models.UserRoleUser,
	}
	err := db.Create(user).Error
	require.NoError(t, err)

	project := &models.Project{
		ID:      uuid.Must(uuid.NewV7()),
		Name:    "Test Project",
		OwnerID: user.ID,
	}
	err = db.Create(project).Error
	require.NoError(t, err)

	t.Run("create asset with valid data", func(t *testing.T) {
		asset := &models.Asset{
			ID:        uuid.Must(uuid.NewV7()),
			Name:      "Test Asset",
			Version:   "1.0.0",
			ProjectID: project.ID,
		}

		err := db.Create(asset).Error
		assert.NoError(t, err)
		assert.Equal(t, "Test Asset", asset.Name)
		assert.Equal(t, "1.0.0", asset.Version)
		assert.Equal(t, project.ID, asset.ProjectID)
	})
}

// TestRepositoryModelValidation tests Repository model validations and
// constraints
func TestRepositoryModelValidation(t *testing.T) {
	db := setupTestDB(t)

	// Create test user and project
	user := &models.User{
		ID:           uuid.Must(uuid.NewV7()),
		Email:        "owner@example.com",
		PasswordHash: "hashed_password",
		Role:         models.UserRoleUser,
	}
	err := db.Create(user).Error
	require.NoError(t, err)

	project := &models.Project{
		ID:      uuid.Must(uuid.NewV7()),
		Name:    "Test Project",
		OwnerID: user.ID,
	}
	err = db.Create(project).Error
	require.NoError(t, err)

	t.Run("create repository with valid data", func(t *testing.T) {
		repo := &models.Repository{
			ID:         uuid.Must(uuid.NewV7()),
			URL:        "https://github.com/example/repo.git",
			Branch:     "main",
			SyncStatus: models.RepositoryStatusPending,
			ProjectID:  project.ID,
		}

		err := db.Create(repo).Error
		assert.NoError(t, err)
		assert.Equal(t, "https://github.com/example/repo.git", repo.URL)
		assert.Equal(t, models.RepositoryStatusPending, repo.SyncStatus)
	})
}

// TestPipelineModelValidation tests Pipeline model validations and constraints
func TestPipelineModelValidation(t *testing.T) {
	db := setupTestDB(t)

	// Create test user and project
	user := &models.User{
		ID:           uuid.Must(uuid.NewV7()),
		Email:        "owner@example.com",
		PasswordHash: "hashed_password",
		Role:         models.UserRoleUser,
	}
	err := db.Create(user).Error
	require.NoError(t, err)

	project := &models.Project{
		ID:      uuid.Must(uuid.NewV7()),
		Name:    "Test Project",
		OwnerID: user.ID,
	}
	err = db.Create(project).Error
	require.NoError(t, err)

	t.Run("create pipeline with valid data", func(t *testing.T) {
		pipeline := &models.Pipeline{
			ID:          uuid.Must(uuid.NewV7()),
			Name:        "Test Pipeline",
			Description: "A test pipeline",
			ProjectID:   project.ID,
		}

		err := db.Create(pipeline).Error
		assert.NoError(t, err)
		assert.Equal(t, "Test Pipeline", pipeline.Name)
		assert.Equal(t, project.ID, pipeline.ProjectID)
	})
}

// TestProjectUserModelValidation tests ProjectUser model validations and
// constraints
func TestProjectUserModelValidation(t *testing.T) {
	db := setupTestDB(t)

	// Create test user and project
	user := &models.User{
		ID:           uuid.Must(uuid.NewV7()),
		Email:        "owner@example.com",
		PasswordHash: "hashed_password",
		Role:         models.UserRoleUser,
	}
	err := db.Create(user).Error
	require.NoError(t, err)

	project := &models.Project{
		ID:      uuid.Must(uuid.NewV7()),
		Name:    "Test Project",
		OwnerID: user.ID,
	}
	err = db.Create(project).Error
	require.NoError(t, err)

	t.Run("create project user with valid data", func(t *testing.T) {
		projectUser := &models.ProjectUser{
			ID:        uuid.Must(uuid.NewV7()),
			ProjectID: project.ID,
			UserID:    user.ID,
			Role:      models.ProjectRoleOwner,
		}

		err := db.Create(projectUser).Error
		assert.NoError(t, err)
		assert.Equal(t, project.ID, projectUser.ProjectID)
		assert.Equal(t, user.ID, projectUser.UserID)
		assert.Equal(t, models.ProjectRoleOwner, projectUser.Role)
	})
}

// TestAssetPipelineModelValidation tests AssetPipeline model validations and
// constraints
func TestAssetPipelineModelValidation(t *testing.T) {
	db := setupTestDB(t)

	// Create test user, project, asset, and pipeline
	user := &models.User{
		ID:           uuid.Must(uuid.NewV7()),
		Email:        "owner@example.com",
		PasswordHash: "hashed_password",
		Role:         models.UserRoleUser,
	}
	err := db.Create(user).Error
	require.NoError(t, err)

	project := &models.Project{
		ID:      uuid.Must(uuid.NewV7()),
		Name:    "Test Project",
		OwnerID: user.ID,
	}
	err = db.Create(project).Error
	require.NoError(t, err)

	asset := &models.Asset{
		ID:        uuid.Must(uuid.NewV7()),
		Name:      "Test Asset",
		Version:   "1.0.0",
		ProjectID: project.ID,
	}
	err = db.Create(asset).Error
	require.NoError(t, err)

	pipeline := &models.Pipeline{
		ID:        uuid.Must(uuid.NewV7()),
		Name:      "Test Pipeline",
		ProjectID: project.ID,
	}
	err = db.Create(pipeline).Error
	require.NoError(t, err)

	t.Run("create asset pipeline with valid data", func(t *testing.T) {
		assetPipeline := &models.AssetPipeline{
			ID:         uuid.Must(uuid.NewV7()),
			PipelineID: pipeline.ID,
			AssetID:    asset.ID,
			Order:      1,
		}

		err := db.Create(assetPipeline).Error
		assert.NoError(t, err)
		assert.Equal(t, pipeline.ID, assetPipeline.PipelineID)
		assert.Equal(t, asset.ID, assetPipeline.AssetID)
		assert.Equal(t, 1, assetPipeline.Order)
	})
}
