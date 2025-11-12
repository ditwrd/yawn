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
package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/ditwrd/yawn/api/internal/domain/models"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// setupTestDB creates an in-memory SQLite database for testing.
func setupAssetTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	// Migrate all models
	err = db.AutoMigrate(
		&models.User{},
		&models.Project{},
		&models.ProjectUser{},
		&models.Asset{},
		&models.Repository{},
		&models.Pipeline{},
		&models.AssetPipeline{},
	)
	require.NoError(t, err)

	return db
}

// createTestAsset creates a test asset with optional relationships.
func createTestAsset(
	t *testing.T,
	db *gorm.DB,
	name, version string,
	projectID, repositoryID *uuid.UUID,
) *models.Asset {
	asset := &models.Asset{
		ID:           uuid.Must(uuid.NewV7()),
		Name:         name,
		Description:  "Test asset description",
		Version:      version,
		ProjectID:    *projectID,
		RepositoryID: repositoryID,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err := db.Create(asset).Error
	require.NoError(t, err)

	return asset
}

// createAssetTestProject creates a test project.
func createAssetTestProject(
	t *testing.T,
	db *gorm.DB,
	name string,
	ownerID *uuid.UUID,
) *models.Project {
	project := &models.Project{
		ID:          uuid.Must(uuid.NewV7()),
		Name:        name,
		Description: "Test project description",
		Visibility:  "private",
		OwnerID:     *ownerID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := db.Create(project).Error
	require.NoError(t, err)

	return project
}

// createAssetTestUser creates a test user.
func createAssetTestUser(t *testing.T, db *gorm.DB, email string) *models.User {
	user := &models.User{
		ID:           uuid.Must(uuid.NewV7()),
		Email:        email,
		PasswordHash: "hashed_password",
		Role:         models.UserRoleUser,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err := db.Create(user).Error
	require.NoError(t, err)

	return user
}

// createAssetTestRepository creates a test repository.
func createAssetTestRepository(
	t *testing.T,
	db *gorm.DB,
	url string,
	projectID *uuid.UUID,
) *models.Repository {
	repo := &models.Repository{
		ID:           uuid.Must(uuid.NewV7()),
		URL:          url,
		Branch:       "main",
		LatestCommit: "abc123",
		SyncStatus:   models.RepositoryStatusSuccess,
		ProjectID:    *projectID,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err := db.Create(repo).Error
	require.NoError(t, err)

	return repo
}

func TestNewAssetRepository(t *testing.T) {
	t.Parallel()

	db := setupAssetTestDB(t)
	repo := NewAssetRepository(db, nil)

	assert.NotNil(t, repo)
	assert.IsType(t, &assetRepository{}, repo)
}

func TestAssetRepository_Create(t *testing.T) {
	t.Parallel()

	db := setupAssetTestDB(t)
	repo := NewAssetRepository(db, nil)
	ctx := context.Background()

	// Create test data
	user := createAssetTestUser(t, db, "test@example.com")
	project := createAssetTestProject(t, db, "Test Project", &user.ID)

	asset := &models.Asset{
		ID:          uuid.Must(uuid.NewV7()),
		Name:        "test-asset",
		Description: "Test asset description",
		Version:     "1.0.0",
		ProjectID:   project.ID,
	}

	// Test create
	err := repo.Create(ctx, asset)
	assert.NoError(t, err)

	// Verify asset was created
	var createdAsset models.Asset

	err = db.First(&createdAsset, "id = ?", asset.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, asset.Name, createdAsset.Name)
	assert.Equal(t, asset.Version, createdAsset.Version)
	assert.Equal(t, project.ID, createdAsset.ProjectID)
}

func TestAssetRepository_GetByID(t *testing.T) {
	t.Parallel()

	db := setupAssetTestDB(t)
	repo := NewAssetRepository(db, nil)
	ctx := context.Background()

	// Create test data
	user := createAssetTestUser(t, db, "test@example.com")
	project := createAssetTestProject(t, db, "Test Project", &user.ID)
	repository := createAssetTestRepository(
		t,
		db,
		"https://github.com/test/repo.git",
		&project.ID,
	)
	asset := createTestAsset(
		t,
		db,
		"test-asset",
		"1.0.0",
		&project.ID,
		&repository.ID,
	)

	// Test get by ID
	retrievedAsset, err := repo.GetByID(ctx, asset.ID.String())
	assert.NoError(t, err)
	assert.NotNil(t, retrievedAsset)
	assert.Equal(t, asset.ID, retrievedAsset.ID)
	assert.Equal(t, asset.Name, retrievedAsset.Name)
	assert.Equal(t, project.ID, retrievedAsset.ProjectID)
	assert.Equal(t, repository.ID, *retrievedAsset.RepositoryID)

	// Test with non-existent ID
	nonExistentID := uuid.Must(uuid.NewV7())
	_, err = repo.GetByID(ctx, nonExistentID.String())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestAssetRepository_GetByProjectID(t *testing.T) {
	t.Parallel()

	db := setupAssetTestDB(t)
	repo := NewAssetRepository(db, nil)
	ctx := context.Background()

	// Create test data
	user := createAssetTestUser(t, db, "test@example.com")
	project1 := createAssetTestProject(t, db, "Project 1", &user.ID)
	project2 := createAssetTestProject(t, db, "Project 2", &user.ID)

	// Create assets in different projects
	asset1 := createTestAsset(t, db, "asset1", "1.0.0", &project1.ID, nil)
	asset2 := createTestAsset(t, db, "asset2", "1.0.0", &project1.ID, nil)
	asset3 := createTestAsset(t, db, "asset3", "1.0.0", &project2.ID, nil)

	// Verify assets were created
	assert.NotNil(t, asset1)
	assert.NotNil(t, asset2)
	assert.NotNil(t, asset3)

	// Test get assets by project ID
	assets, err := repo.GetByProjectID(ctx, project1.ID.String(), 10, 0)
	assert.NoError(t, err)
	assert.Len(t, assets, 2)

	assetNames := make([]string, len(assets))
	for i, asset := range assets {
		assetNames[i] = asset.Name
	}

	assert.Contains(t, assetNames, "asset1")
	assert.Contains(t, assetNames, "asset2")

	// Test with pagination
	assets, err = repo.GetByProjectID(ctx, project1.ID.String(), 1, 0)
	assert.NoError(t, err)
	assert.Len(t, assets, 1)

	assets, err = repo.GetByProjectID(ctx, project1.ID.String(), 1, 1)
	assert.NoError(t, err)
	assert.Len(t, assets, 1)
}

func TestAssetRepository_GetByRepositoryID(t *testing.T) {
	t.Parallel()

	db := setupAssetTestDB(t)
	repo := NewAssetRepository(db, nil)
	ctx := context.Background()

	// Create test data
	user := createAssetTestUser(t, db, "test@example.com")
	project := createAssetTestProject(t, db, "Test Project", &user.ID)
	repository := createAssetTestRepository(
		t,
		db,
		"https://github.com/test/repo.git",
		&project.ID,
	)

	// Create assets with and without repository
	asset1 := createTestAsset(
		t,
		db,
		"asset1",
		"1.0.0",
		&project.ID,
		&repository.ID,
	)
	asset2 := createTestAsset(
		t,
		db,
		"asset2",
		"1.0.0",
		&project.ID,
		&repository.ID,
	)
	asset3 := createTestAsset(t, db, "asset3", "1.0.0", &project.ID, nil)

	// Verify assets were created
	assert.NotNil(t, asset1)
	assert.NotNil(t, asset2)
	assert.NotNil(t, asset3)

	// Test get assets by repository ID
	assets, err := repo.GetByRepositoryID(ctx, repository.ID.String())
	assert.NoError(t, err)
	assert.Len(t, assets, 2)

	assetNames := make([]string, len(assets))
	for i, asset := range assets {
		assetNames[i] = asset.Name
	}

	assert.Contains(t, assetNames, "asset1")
	assert.Contains(t, assetNames, "asset2")
	assert.NotContains(t, assetNames, "asset3")
}

func TestAssetRepository_List(t *testing.T) {
	t.Parallel()

	db := setupAssetTestDB(t)
	repo := NewAssetRepository(db, nil)
	ctx := context.Background()

	// Create test data
	user := createAssetTestUser(t, db, "test@example.com")
	project1 := createAssetTestProject(t, db, "Project 1", &user.ID)
	project2 := createAssetTestProject(t, db, "Project 2", &user.ID)
	repository := createAssetTestRepository(
		t,
		db,
		"https://github.com/test/repo.git",
		&project1.ID,
	)

	// Create various assets
	asset1 := createTestAsset(
		t,
		db,
		"test-asset",
		"1.0.0",
		&project1.ID,
		&repository.ID,
	)
	asset2 := createTestAsset(
		t,
		db,
		"production-asset",
		"2.0.0",
		&project1.ID,
		nil,
	)
	asset3 := createTestAsset(t, db, "test-asset", "1.1.0", &project2.ID, nil)

	// Verify assets were created
	assert.NotNil(t, asset1)
	assert.NotNil(t, asset2)
	assert.NotNil(t, asset3)

	tests := []struct {
		name    string
		filters AssetFilters
		wantLen int
	}{
		{
			name:    "list all assets",
			filters: AssetFilters{},
			wantLen: 3,
		},
		{
			name: "filter by project ID",
			filters: AssetFilters{
				ProjectID: project1.ID.String(),
			},
			wantLen: 2,
		},
		{
			name: "filter by repository ID",
			filters: AssetFilters{
				RepositoryID: repository.ID.String(),
			},
			wantLen: 1,
		},
		{
			name: "filter by name",
			filters: AssetFilters{
				Name: "test-asset",
			},
			wantLen: 2,
		},
		{
			name: "filter by version",
			filters: AssetFilters{
				Version: "1.0.0",
			},
			wantLen: 1,
		},
		{
			name: "search across name and description",
			filters: AssetFilters{
				Search: "test",
			},
			wantLen: 3,
		},
		{
			name: "filter by multiple criteria",
			filters: AssetFilters{
				ProjectID: project1.ID.String(),
				Name:      "test-asset",
			},
			wantLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assets, err := repo.List(ctx, 100, 0, tt.filters)
			assert.NoError(t, err)
			assert.Len(t, assets, tt.wantLen)
		})
	}

	// Test pagination
	assets, err := repo.List(ctx, 2, 0, AssetFilters{})
	assert.NoError(t, err)
	assert.Len(t, assets, 2)

	assets, err = repo.List(ctx, 2, 2, AssetFilters{})
	assert.NoError(t, err)
	assert.Len(t, assets, 1)
}

func TestAssetRepository_Update(t *testing.T) {
	t.Parallel()

	db := setupAssetTestDB(t)
	repo := NewAssetRepository(db, nil)
	ctx := context.Background()

	// Create test data
	user := createAssetTestUser(t, db, "test@example.com")
	project := createAssetTestProject(t, db, "Test Project", &user.ID)
	asset := createTestAsset(t, db, "test-asset", "1.0.0", &project.ID, nil)

	// Update asset
	asset.Name = "updated-asset"
	asset.Description = "Updated description"
	asset.Version = "2.0.0"

	err := repo.Update(ctx, asset)
	assert.NoError(t, err)

	// Verify update
	var updatedAsset models.Asset

	err = db.First(&updatedAsset, "id = ?", asset.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, "updated-asset", updatedAsset.Name)
	assert.Equal(t, "Updated description", updatedAsset.Description)
	assert.Equal(t, "2.0.0", updatedAsset.Version)
}

func TestAssetRepository_Delete(t *testing.T) {
	t.Parallel()

	db := setupAssetTestDB(t)
	repo := NewAssetRepository(db, nil)
	ctx := context.Background()

	// Create test data
	user := createAssetTestUser(t, db, "test@example.com")
	project := createAssetTestProject(t, db, "Test Project", &user.ID)
	asset := createTestAsset(t, db, "test-asset", "1.0.0", &project.ID, nil)

	// Delete asset
	err := repo.Delete(ctx, asset.ID.String())
	assert.NoError(t, err)

	// Verify soft delete (asset should not be found in normal queries)
	var deletedAsset models.Asset

	err = db.First(&deletedAsset, "id = ?", asset.ID).Error
	assert.Error(t, err)
	assert.Equal(t, gorm.ErrRecordNotFound, err)

	// But should be found with Unscoped
	err = db.Unscoped().First(&deletedAsset, "id = ?", asset.ID).Error
	assert.NoError(t, err)
	assert.NotNil(t, deletedAsset.DeletedAt)
}

func TestAssetRepository_Search(t *testing.T) {
	t.Parallel()

	db := setupAssetTestDB(t)
	repo := NewAssetRepository(db, nil)
	ctx := context.Background()

	// Create test data
	user := createAssetTestUser(t, db, "test@example.com")
	project := createAssetTestProject(t, db, "Test Project", &user.ID)

	asset1 := createTestAsset(
		t,
		db,
		"production-database",
		"1.0.0",
		&project.ID,
		nil,
	)
	asset2 := createTestAsset(t, db, "test-database", "1.0.0", &project.ID, nil)
	asset3 := createTestAsset(t, db, "web-server", "1.0.0", &project.ID, nil)

	// Update descriptions for search testing
	db.Model(&asset1).Update("description", "Production database server")
	db.Model(&asset2).Update("description", "Test database instance")
	db.Model(&asset3).Update("description", "Web application server")

	tests := []struct {
		name    string
		query   string
		wantLen int
	}{
		{
			name:    "search for 'database'",
			query:   "database",
			wantLen: 2,
		},
		{
			name:    "search for 'server'",
			query:   "server",
			wantLen: 2,
		},
		{
			name:    "search for 'test'",
			query:   "test",
			wantLen: 1,
		},
		{
			name:    "search for 'production'",
			query:   "production",
			wantLen: 1,
		},
		{
			name:    "search for non-existent term",
			query:   "nonexistent",
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assets, err := repo.Search(ctx, tt.query, 100, 0)
			assert.NoError(t, err)
			assert.Len(t, assets, tt.wantLen)
		})
	}
}

func TestAssetRepository_GetVersionHistory(t *testing.T) {
	t.Parallel()

	db := setupAssetTestDB(t)
	repo := NewAssetRepository(db, nil)
	ctx := context.Background()

	// Create test data
	user := createAssetTestUser(t, db, "test@example.com")
	project := createAssetTestProject(t, db, "Test Project", &user.ID)

	// Create multiple versions of the same asset
	asset1 := createTestAsset(t, db, "test-asset", "1.0.0", &project.ID, nil)
	asset2 := createTestAsset(t, db, "test-asset", "1.1.0", &project.ID, nil)
	asset3 := createTestAsset(t, db, "test-asset", "2.0.0", &project.ID, nil)
	otherAsset := createTestAsset(t, db, "other-asset", "1.0.0", &project.ID, nil)

	// Verify assets were created
	assert.NotNil(t, asset1)
	assert.NotNil(t, asset2)
	assert.NotNil(t, asset3)
	assert.NotNil(t, otherAsset)

	// Get version history
	versions, err := repo.GetVersionHistory(
		ctx,
		project.ID.String(),
		"test-asset",
	)
	assert.NoError(t, err)
	assert.Len(t, versions, 3)

	// Verify order (should be descending by creation time)
	assert.Equal(t, "2.0.0", versions[0].Version)
	assert.Equal(t, "1.1.0", versions[1].Version)
	assert.Equal(t, "1.0.0", versions[2].Version)

	// Test non-existent asset
	versions, err = repo.GetVersionHistory(
		ctx,
		project.ID.String(),
		"non-existent",
	)
	assert.NoError(t, err)
	assert.Empty(t, versions)
}

func TestAssetRepository_GetLatestVersion(t *testing.T) {
	t.Parallel()

	db := setupAssetTestDB(t)
	repo := NewAssetRepository(db, nil)
	ctx := context.Background()

	// Create test data
	user := createAssetTestUser(t, db, "test@example.com")
	project := createAssetTestProject(t, db, "Test Project", &user.ID)

	// Create multiple versions of the same asset
	createTestAsset(t, db, "test-asset", "1.0.0", &project.ID, nil)
	_ = createTestAsset(t, db, "test-asset", "1.1.0", &project.ID, nil)
	createTestAsset(t, db, "test-asset", "2.0.0", &project.ID, nil)

	// Get latest version
	latest, err := repo.GetLatestVersion(ctx, project.ID.String(), "test-asset")
	assert.NoError(t, err)
	assert.NotNil(t, latest)
	assert.Equal(t, "2.0.0", latest.Version)

	// Test non-existent asset
	_, err = repo.GetLatestVersion(ctx, project.ID.String(), "non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no versions found")
}

func TestAssetRepository_Count(t *testing.T) {
	t.Parallel()

	db := setupAssetTestDB(t)
	repo := NewAssetRepository(db, nil)
	ctx := context.Background()

	// Create test data
	user := createAssetTestUser(t, db, "test@example.com")
	project1 := createAssetTestProject(t, db, "Project 1", &user.ID)
	project2 := createAssetTestProject(t, db, "Project 2", &user.ID)

	// Create assets
	createTestAsset(t, db, "asset1", "1.0.0", &project1.ID, nil)
	createTestAsset(t, db, "asset2", "1.0.0", &project1.ID, nil)
	createTestAsset(t, db, "asset3", "1.0.0", &project2.ID, nil)

	tests := []struct {
		name    string
		filters AssetFilters
		want    int64
	}{
		{
			name:    "count all assets",
			filters: AssetFilters{},
			want:    3,
		},
		{
			name: "count by project ID",
			filters: AssetFilters{
				ProjectID: project1.ID.String(),
			},
			want: 2,
		},
		{
			name: "count by name",
			filters: AssetFilters{
				Name: "asset1",
			},
			want: 1,
		},
		{
			name: "count by version",
			filters: AssetFilters{
				Version: "1.0.0",
			},
			want: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count, err := repo.Count(ctx, tt.filters)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, count)
		})
	}
}

func TestAssetRepository_Exists(t *testing.T) {
	t.Parallel()

	db := setupAssetTestDB(t)
	repo := NewAssetRepository(db, nil)
	ctx := context.Background()

	// Create test data
	user := createAssetTestUser(t, db, "test@example.com")
	project := createAssetTestProject(t, db, "Test Project", &user.ID)
	asset := createTestAsset(t, db, "test-asset", "1.0.0", &project.ID, nil)

	// Test existing asset
	exists, err := repo.Exists(ctx, asset.ID.String())
	assert.NoError(t, err)
	assert.True(t, exists)

	// Test non-existent asset
	nonExistentID := uuid.Must(uuid.NewV7())
	exists, err = repo.Exists(ctx, nonExistentID.String())
	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestAssetRepository_ExistsByName(t *testing.T) {
	t.Parallel()

	db := setupAssetTestDB(t)
	repo := NewAssetRepository(db, nil)
	ctx := context.Background()

	// Create test data
	user := createAssetTestUser(t, db, "test@example.com")
	project := createAssetTestProject(t, db, "Test Project", &user.ID)
	createTestAsset(t, db, "test-asset", "1.0.0", &project.ID, nil)

	// Test existing asset name
	exists, err := repo.ExistsByName(ctx, project.ID.String(), "test-asset")
	assert.NoError(t, err)
	assert.True(t, exists)

	// Test non-existent asset name
	exists, err = repo.ExistsByName(ctx, project.ID.String(), "non-existent")
	assert.NoError(t, err)
	assert.False(t, exists)

	// Test with different project
	project2 := createAssetTestProject(t, db, "Project 2", &user.ID)
	exists, err = repo.ExistsByName(ctx, project2.ID.String(), "test-asset")
	assert.NoError(t, err)
	assert.False(t, exists)
}
