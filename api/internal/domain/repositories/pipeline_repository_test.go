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

// setupPipelineTestDB creates an in-memory SQLite database for testing.
func setupPipelineTestDB(t *testing.T) *gorm.DB {
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
		&models.PipelineExecution{},
		&models.ExecutionStep{},
		&models.PipelineDependency{},
		&models.AssetPipeline{},
	)
	require.NoError(t, err)

	return db
}

// createTestPipeline creates a test pipeline with optional relationships.
func createTestPipeline(
	t *testing.T,
	db *gorm.DB,
	name string,
	projectID *uuid.UUID,
) *models.Pipeline {
	pipeline := &models.Pipeline{
		ID:          uuid.Must(uuid.NewV7()),
		Name:        name,
		Description: "Test pipeline description",
		ProjectID:   *projectID,
		Status:      models.PipelineStatusDraft,
		Config:      "{}",
		IsEnabled:   true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := db.Create(pipeline).Error
	require.NoError(t, err)

	return pipeline
}

// createPipelineTestProject creates a test project.
func createPipelineTestProject(
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

// createPipelineTestUser creates a test user.
func createPipelineTestUser(
	t *testing.T,
	db *gorm.DB,
	email string,
) *models.User {
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

// createTestPipelineExecution creates a test pipeline execution.
func createTestPipelineExecution(
	t *testing.T,
	db *gorm.DB,
	pipelineID *uuid.UUID,
	status models.ExecutionStatus,
) *models.PipelineExecution {
	execution := &models.PipelineExecution{
		ID:          uuid.Must(uuid.NewV7()),
		PipelineID:  *pipelineID,
		Status:      status,
		TriggerType: "manual",
		Duration:    0,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if status == models.ExecutionStatusRunning {
		startedAt := time.Now()
		execution.StartedAt = &startedAt
	}

	if status == models.ExecutionStatusCompleted ||
		status == models.ExecutionStatusFailed {
		startedAt := time.Now().Add(-5 * time.Minute)
		completedAt := time.Now()
		execution.StartedAt = &startedAt
		execution.CompletedAt = &completedAt
		execution.Duration = int(completedAt.Sub(startedAt).Seconds())
	}

	err := db.Create(execution).Error
	require.NoError(t, err)

	return execution
}

func TestNewPipelineRepository(t *testing.T) {
	t.Parallel()

	db := setupPipelineTestDB(t)
	repo := NewPipelineRepository(db, nil)

	assert.NotNil(t, repo)
	assert.IsType(t, &pipelineRepository{}, repo)
}

func TestPipelineRepository_Create(t *testing.T) {
	t.Parallel()

	db := setupPipelineTestDB(t)
	repo := NewPipelineRepository(db, nil)
	ctx := context.Background()

	// Create test data
	user := createPipelineTestUser(t, db, "test@example.com")
	project := createPipelineTestProject(t, db, "Test Project", &user.ID)

	pipeline := &models.Pipeline{
		ID:          uuid.Must(uuid.NewV7()),
		Name:        "test-pipeline",
		Description: "Test pipeline description",
		ProjectID:   project.ID,
		Status:      models.PipelineStatusDraft,
		Config:      "{}",
		IsEnabled:   true,
	}

	// Test create
	err := repo.Create(ctx, pipeline)
	assert.NoError(t, err)

	// Verify pipeline was created
	var createdPipeline models.Pipeline

	err = db.First(&createdPipeline, "id = ?", pipeline.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, pipeline.Name, createdPipeline.Name)
	assert.Equal(t, pipeline.Status, createdPipeline.Status)
	assert.Equal(t, project.ID, createdPipeline.ProjectID)
}

func TestPipelineRepository_GetByID(t *testing.T) {
	t.Parallel()

	db := setupPipelineTestDB(t)
	repo := NewPipelineRepository(db, nil)
	ctx := context.Background()

	// Create test data
	user := createPipelineTestUser(t, db, "test@example.com")
	project := createPipelineTestProject(t, db, "Test Project", &user.ID)
	pipeline := createTestPipeline(t, db, "test-pipeline", &project.ID)

	// Test get by ID
	retrievedPipeline, err := repo.GetByID(ctx, pipeline.ID.String())
	assert.NoError(t, err)
	assert.NotNil(t, retrievedPipeline)
	assert.Equal(t, pipeline.ID, retrievedPipeline.ID)
	assert.Equal(t, pipeline.Name, retrievedPipeline.Name)
	assert.Equal(t, project.ID, retrievedPipeline.ProjectID)

	// Test with non-existent ID
	nonExistentID := uuid.Must(uuid.NewV7())
	_, err = repo.GetByID(ctx, nonExistentID.String())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestPipelineRepository_GetByProjectID(t *testing.T) {
	t.Parallel()

	db := setupPipelineTestDB(t)
	repo := NewPipelineRepository(db, nil)
	ctx := context.Background()

	// Create test data
	user := createPipelineTestUser(t, db, "test@example.com")
	project1 := createPipelineTestProject(t, db, "Project 1", &user.ID)
	project2 := createPipelineTestProject(t, db, "Project 2", &user.ID)

	// Create pipelines in different projects
	_ = createTestPipeline(t, db, "pipeline1", &project1.ID)
	_ = createTestPipeline(t, db, "pipeline2", &project1.ID)
	_ = createTestPipeline(t, db, "pipeline3", &project2.ID)

	// Test get pipelines by project ID
	pipelines, err := repo.GetByProjectID(ctx, project1.ID.String(), 10, 0)
	assert.NoError(t, err)
	assert.Len(t, pipelines, 2)

	pipelineNames := make([]string, len(pipelines))
	for i, pipeline := range pipelines {
		pipelineNames[i] = pipeline.Name
	}

	assert.Contains(t, pipelineNames, "pipeline1")
	assert.Contains(t, pipelineNames, "pipeline2")

	// Test with pagination
	pipelines, err = repo.GetByProjectID(ctx, project1.ID.String(), 1, 0)
	assert.NoError(t, err)
	assert.Len(t, pipelines, 1)

	pipelines, err = repo.GetByProjectID(ctx, project1.ID.String(), 1, 1)
	assert.NoError(t, err)
	assert.Len(t, pipelines, 1)
}

func TestPipelineRepository_List(t *testing.T) {
	t.Parallel()

	db := setupPipelineTestDB(t)
	repo := NewPipelineRepository(db, nil)
	ctx := context.Background()

	// Create test data
	user := createPipelineTestUser(t, db, "test@example.com")
	project1 := createPipelineTestProject(t, db, "Project 1", &user.ID)
	project2 := createPipelineTestProject(t, db, "Project 2", &user.ID)

	// Create various pipelines
	_ = createTestPipeline(t, db, "test-pipeline", &project1.ID)
	pipeline2 := createTestPipeline(t, db, "production-pipeline", &project1.ID)
	pipeline3 := createTestPipeline(t, db, "test-pipeline", &project2.ID)

	// Update pipeline statuses
	db.Model(&pipeline2).Update("status", models.PipelineStatusActive)
	db.Model(&pipeline3).Update("status", models.PipelineStatusActive)
	db.Model(&pipeline3).Update("is_enabled", false)

	tests := []struct {
		name    string
		filters PipelineFilters
		wantLen int
	}{
		{
			name:    "list all pipelines",
			filters: PipelineFilters{},
			wantLen: 3,
		},
		{
			name: "filter by project ID",
			filters: PipelineFilters{
				ProjectID: project1.ID.String(),
			},
			wantLen: 2,
		},
		{
			name: "filter by status",
			filters: PipelineFilters{
				Status: models.PipelineStatusActive,
			},
			wantLen: 2,
		},
		{
			name: "filter by name",
			filters: PipelineFilters{
				Name: "test-pipeline",
			},
			wantLen: 2,
		},
		{
			name: "filter by enabled status",
			filters: PipelineFilters{
				IsEnabled: func() *bool {
					b := true

					return &b
				}(),
			},
			wantLen: 2,
		},
		{
			name: "search across name and description",
			filters: PipelineFilters{
				Search: "test",
			},
			wantLen: 3,
		},
		{
			name: "filter by multiple criteria",
			filters: PipelineFilters{
				ProjectID: project1.ID.String(),
				Status:    models.PipelineStatusDraft,
			},
			wantLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pipelines, err := repo.List(ctx, 100, 0, tt.filters)
			assert.NoError(t, err)
			assert.Len(t, pipelines, tt.wantLen)
		})
	}

	// Test pagination
	pipelines, err := repo.List(ctx, 2, 0, PipelineFilters{})
	assert.NoError(t, err)
	assert.Len(t, pipelines, 2)

	pipelines, err = repo.List(ctx, 2, 2, PipelineFilters{})
	assert.NoError(t, err)
	assert.Len(t, pipelines, 1)
}

func TestPipelineRepository_Update(t *testing.T) {
	t.Parallel()

	db := setupPipelineTestDB(t)
	repo := NewPipelineRepository(db, nil)
	ctx := context.Background()

	// Create test data
	user := createPipelineTestUser(t, db, "test@example.com")
	project := createPipelineTestProject(t, db, "Test Project", &user.ID)
	pipeline := createTestPipeline(t, db, "test-pipeline", &project.ID)

	// Update pipeline
	pipeline.Name = "updated-pipeline"
	pipeline.Description = "Updated description"
	pipeline.Status = models.PipelineStatusActive
	pipeline.IsEnabled = false

	err := repo.Update(ctx, pipeline)
	assert.NoError(t, err)

	// Verify update
	var updatedPipeline models.Pipeline

	err = db.First(&updatedPipeline, "id = ?", pipeline.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, "updated-pipeline", updatedPipeline.Name)
	assert.Equal(t, "Updated description", updatedPipeline.Description)
	assert.Equal(t, models.PipelineStatusActive, updatedPipeline.Status)
	assert.False(t, updatedPipeline.IsEnabled)
}

func TestPipelineRepository_Delete(t *testing.T) {
	t.Parallel()

	db := setupPipelineTestDB(t)
	repo := NewPipelineRepository(db, nil)
	ctx := context.Background()

	// Create test data
	user := createPipelineTestUser(t, db, "test@example.com")
	project := createPipelineTestProject(t, db, "Test Project", &user.ID)
	pipeline := createTestPipeline(t, db, "test-pipeline", &project.ID)

	// Delete pipeline
	err := repo.Delete(ctx, pipeline.ID.String())
	assert.NoError(t, err)

	// Verify soft delete (pipeline should not be found in normal queries)
	var deletedPipeline models.Pipeline

	err = db.First(&deletedPipeline, "id = ?", pipeline.ID).Error
	assert.Error(t, err)
	assert.Equal(t, gorm.ErrRecordNotFound, err)

	// But should be found with Unscoped
	err = db.Unscoped().First(&deletedPipeline, "id = ?", pipeline.ID).Error
	assert.NoError(t, err)
	assert.NotNil(t, deletedPipeline.DeletedAt)
}

func TestPipelineRepository_Search(t *testing.T) {
	t.Parallel()

	db := setupPipelineTestDB(t)
	repo := NewPipelineRepository(db, nil)
	ctx := context.Background()

	// Create test data
	user := createPipelineTestUser(t, db, "test@example.com")
	project := createPipelineTestProject(t, db, "Test Project", &user.ID)

	pipeline1 := createTestPipeline(t, db, "production-pipeline", &project.ID)
	pipeline2 := createTestPipeline(t, db, "test-pipeline", &project.ID)
	pipeline3 := createTestPipeline(t, db, "web-pipeline", &project.ID)

	// Update descriptions for search testing
	db.Model(&pipeline1).Update("description", "Production data pipeline")
	db.Model(&pipeline2).Update("description", "Test execution pipeline")
	db.Model(&pipeline3).Update("description", "Web application pipeline")

	tests := []struct {
		name    string
		query   string
		wantLen int
	}{
		{
			name:    "search for 'pipeline'",
			query:   "pipeline",
			wantLen: 3,
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
			name:    "search for 'web'",
			query:   "web",
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
			pipelines, err := repo.Search(ctx, tt.query, 100, 0)
			assert.NoError(t, err)
			assert.Len(t, pipelines, tt.wantLen)
		})
	}
}

func TestPipelineRepository_GetActivePipelines(t *testing.T) {
	t.Parallel()

	db := setupPipelineTestDB(t)
	repo := NewPipelineRepository(db, nil)
	ctx := context.Background()

	// Create test data
	user := createPipelineTestUser(t, db, "test@example.com")
	project := createPipelineTestProject(t, db, "Test Project", &user.ID)

	// Create pipelines with different statuses
	pipeline1 := createTestPipeline(t, db, "active-pipeline", &project.ID)
	pipeline2 := createTestPipeline(t, db, "disabled-pipeline", &project.ID)
	_ = createTestPipeline(t, db, "draft-pipeline", &project.ID)

	// Update statuses and enabled flags
	db.Model(&pipeline1).Update("status", models.PipelineStatusActive)
	db.Model(&pipeline2).Update("status", models.PipelineStatusActive)
	db.Model(&pipeline2).Update("is_enabled", false)
	// pipeline3 remains draft

	// Get active pipelines
	pipelines, err := repo.GetActivePipelines(ctx)
	assert.NoError(t, err)
	assert.Len(t, pipelines, 1) // Only pipeline1 should be active and enabled

	assert.Equal(t, "active-pipeline", pipelines[0].Name)
}

func TestPipelineRepository_UpdateStatus(t *testing.T) {
	t.Parallel()

	db := setupPipelineTestDB(t)
	repo := NewPipelineRepository(db, nil)
	ctx := context.Background()

	// Create test data
	user := createPipelineTestUser(t, db, "test@example.com")
	project := createPipelineTestProject(t, db, "Test Project", &user.ID)
	pipeline := createTestPipeline(t, db, "test-pipeline", &project.ID)

	// Update status
	err := repo.UpdateStatus(
		ctx,
		pipeline.ID.String(),
		models.PipelineStatusActive,
	)
	assert.NoError(t, err)

	// Verify update
	var updatedPipeline models.Pipeline

	err = db.First(&updatedPipeline, "id = ?", pipeline.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, models.PipelineStatusActive, updatedPipeline.Status)
}

func TestPipelineRepository_Count(t *testing.T) {
	t.Parallel()

	db := setupPipelineTestDB(t)
	repo := NewPipelineRepository(db, nil)
	ctx := context.Background()

	// Create test data
	user := createPipelineTestUser(t, db, "test@example.com")
	project1 := createPipelineTestProject(t, db, "Project 1", &user.ID)
	project2 := createPipelineTestProject(t, db, "Project 2", &user.ID)

	// Create pipelines
	createTestPipeline(t, db, "pipeline1", &project1.ID)
	createTestPipeline(t, db, "pipeline2", &project1.ID)
	createTestPipeline(t, db, "pipeline3", &project2.ID)

	tests := []struct {
		name    string
		filters PipelineFilters
		want    int64
	}{
		{
			name:    "count all pipelines",
			filters: PipelineFilters{},
			want:    3,
		},
		{
			name: "count by project ID",
			filters: PipelineFilters{
				ProjectID: project1.ID.String(),
			},
			want: 2,
		},
		{
			name: "count by name",
			filters: PipelineFilters{
				Name: "pipeline1",
			},
			want: 1,
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

func TestPipelineRepository_Exists(t *testing.T) {
	t.Parallel()

	db := setupPipelineTestDB(t)
	repo := NewPipelineRepository(db, nil)
	ctx := context.Background()

	// Create test data
	user := createPipelineTestUser(t, db, "test@example.com")
	project := createPipelineTestProject(t, db, "Test Project", &user.ID)
	pipeline := createTestPipeline(t, db, "test-pipeline", &project.ID)

	// Test existing pipeline
	exists, err := repo.Exists(ctx, pipeline.ID.String())
	assert.NoError(t, err)
	assert.True(t, exists)

	// Test non-existent pipeline
	nonExistentID := uuid.Must(uuid.NewV7())
	exists, err = repo.Exists(ctx, nonExistentID.String())
	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestPipelineRepository_ExistsByName(t *testing.T) {
	t.Parallel()

	db := setupPipelineTestDB(t)
	repo := NewPipelineRepository(db, nil)
	ctx := context.Background()

	// Create test data
	user := createPipelineTestUser(t, db, "test@example.com")
	project := createPipelineTestProject(t, db, "Test Project", &user.ID)
	createTestPipeline(t, db, "test-pipeline", &project.ID)

	// Test existing pipeline name
	exists, err := repo.ExistsByName(ctx, project.ID.String(), "test-pipeline")
	assert.NoError(t, err)
	assert.True(t, exists)

	// Test non-existent pipeline name
	exists, err = repo.ExistsByName(ctx, project.ID.String(), "non-existent")
	assert.NoError(t, err)
	assert.False(t, exists)

	// Test with different project
	project2 := createPipelineTestProject(t, db, "Project 2", &user.ID)
	exists, err = repo.ExistsByName(ctx, project2.ID.String(), "test-pipeline")
	assert.NoError(t, err)
	assert.False(t, exists)
}

// === Pipeline Execution Tests ===

func TestPipelineRepository_CreateExecution(t *testing.T) {
	t.Parallel()

	db := setupPipelineTestDB(t)
	repo := NewPipelineRepository(db, nil)
	ctx := context.Background()

	// Create test data
	user := createPipelineTestUser(t, db, "test@example.com")
	project := createPipelineTestProject(t, db, "Test Project", &user.ID)
	pipeline := createTestPipeline(t, db, "test-pipeline", &project.ID)

	execution := &models.PipelineExecution{
		ID:          uuid.Must(uuid.NewV7()),
		PipelineID:  pipeline.ID,
		Status:      models.ExecutionStatusPending,
		TriggerType: "manual",
		Duration:    0,
		Config:      "{}",
	}

	// Test create
	err := repo.CreateExecution(ctx, execution)
	assert.NoError(t, err)

	// Verify execution was created
	var createdExecution models.PipelineExecution

	err = db.First(&createdExecution, "id = ?", execution.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, execution.PipelineID, createdExecution.PipelineID)
	assert.Equal(t, execution.Status, createdExecution.Status)
}

func TestPipelineRepository_GetExecutionByID(t *testing.T) {
	t.Parallel()

	db := setupPipelineTestDB(t)
	repo := NewPipelineRepository(db, nil)
	ctx := context.Background()

	// Create test data
	user := createPipelineTestUser(t, db, "test@example.com")
	project := createPipelineTestProject(t, db, "Test Project", &user.ID)
	pipeline := createTestPipeline(t, db, "test-pipeline", &project.ID)
	execution := createTestPipelineExecution(
		t,
		db,
		&pipeline.ID,
		models.ExecutionStatusCompleted,
	)

	// Test get by ID
	retrievedExecution, err := repo.GetExecutionByID(ctx, execution.ID.String())
	assert.NoError(t, err)
	assert.NotNil(t, retrievedExecution)
	assert.Equal(t, execution.ID, retrievedExecution.ID)
	assert.Equal(t, execution.PipelineID, retrievedExecution.PipelineID)
	assert.Equal(t, execution.Status, retrievedExecution.Status)

	// Test with non-existent ID
	nonExistentID := uuid.Must(uuid.NewV7())
	_, err = repo.GetExecutionByID(ctx, nonExistentID.String())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestPipelineRepository_GetExecutionsByPipelineID(t *testing.T) {
	t.Parallel()

	db := setupPipelineTestDB(t)
	repo := NewPipelineRepository(db, nil)
	ctx := context.Background()

	// Create test data
	user := createPipelineTestUser(t, db, "test@example.com")
	project := createPipelineTestProject(t, db, "Test Project", &user.ID)
	pipeline := createTestPipeline(t, db, "test-pipeline", &project.ID)

	// Create multiple executions
	execution1 := createTestPipelineExecution(
		t,
		db,
		&pipeline.ID,
		models.ExecutionStatusCompleted,
	)
	execution2 := createTestPipelineExecution(
		t,
		db,
		&pipeline.ID,
		models.ExecutionStatusFailed,
	)
	execution3 := createTestPipelineExecution(
		t,
		db,
		&pipeline.ID,
		models.ExecutionStatusRunning,
	)

	// Get executions by pipeline ID
	executions, err := repo.GetExecutionsByPipelineID(
		ctx,
		pipeline.ID.String(),
		10,
		0,
	)
	assert.NoError(t, err)
	assert.Len(t, executions, 3)

	// Verify order (should be descending by creation time)
	assert.Equal(t, execution3.ID, executions[0].ID) // Most recent
	assert.Equal(t, execution2.ID, executions[1].ID)
	assert.Equal(t, execution1.ID, executions[2].ID) // Oldest
}

func TestPipelineRepository_UpdateExecutionStatus(t *testing.T) {
	t.Parallel()

	db := setupPipelineTestDB(t)
	repo := NewPipelineRepository(db, nil)
	ctx := context.Background()

	// Create test data
	user := createPipelineTestUser(t, db, "test@example.com")
	project := createPipelineTestProject(t, db, "Test Project", &user.ID)
	pipeline := createTestPipeline(t, db, "test-pipeline", &project.ID)
	execution := createTestPipelineExecution(
		t,
		db,
		&pipeline.ID,
		models.ExecutionStatusPending,
	)

	// Update status to running
	err := repo.UpdateExecutionStatus(
		ctx,
		execution.ID.String(),
		models.ExecutionStatusRunning,
	)
	assert.NoError(t, err)

	// Verify status update and started_at field
	var updatedExecution models.PipelineExecution

	err = db.First(&updatedExecution, "id = ?", execution.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, models.ExecutionStatusRunning, updatedExecution.Status)
	assert.NotNil(t, updatedExecution.StartedAt)

	// Wait a bit to ensure duration > 0
	time.Sleep(10 * time.Millisecond)

	// Update status to completed
	err = repo.UpdateExecutionStatus(
		ctx,
		execution.ID.String(),
		models.ExecutionStatusCompleted,
	)
	assert.NoError(t, err)

	// Verify completion fields
	err = db.First(&updatedExecution, "id = ?", execution.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, models.ExecutionStatusCompleted, updatedExecution.Status)
	assert.NotNil(t, updatedExecution.CompletedAt)
	assert.Positive(t, updatedExecution.Duration)
}

func TestPipelineRepository_GetRunningExecutions(t *testing.T) {
	t.Parallel()

	db := setupPipelineTestDB(t)
	repo := NewPipelineRepository(db, nil)
	ctx := context.Background()

	// Create test data
	user := createPipelineTestUser(t, db, "test@example.com")
	project := createPipelineTestProject(t, db, "Test Project", &user.ID)
	pipeline1 := createTestPipeline(t, db, "pipeline1", &project.ID)
	pipeline2 := createTestPipeline(t, db, "pipeline2", &project.ID)

	// Create executions with different statuses
	execution1 := createTestPipelineExecution(
		t,
		db,
		&pipeline1.ID,
		models.ExecutionStatusRunning,
	)
	_ = createTestPipelineExecution(
		t,
		db,
		&pipeline2.ID,
		models.ExecutionStatusCompleted,
	)
	createTestPipelineExecution(
		t,
		db,
		&pipeline1.ID,
		models.ExecutionStatusFailed,
	)

	// Get running executions
	executions, err := repo.GetRunningExecutions(ctx)
	assert.NoError(t, err)
	assert.Len(t, executions, 1)
	assert.Equal(t, execution1.ID, executions[0].ID)
}

// === Dependency Tests ===

func TestPipelineRepository_CreateDependency(t *testing.T) {
	t.Parallel()

	db := setupPipelineTestDB(t)
	repo := NewPipelineRepository(db, nil)
	ctx := context.Background()

	// Create test data
	user := createPipelineTestUser(t, db, "test@example.com")
	project := createPipelineTestProject(t, db, "Test Project", &user.ID)
	pipeline1 := createTestPipeline(t, db, "pipeline1", &project.ID)
	pipeline2 := createTestPipeline(t, db, "pipeline2", &project.ID)

	dependency := &models.PipelineDependency{
		ID:                  uuid.Must(uuid.NewV7()),
		PipelineID:          pipeline1.ID,
		DependsOnPipelineID: pipeline2.ID,
		Condition:           "success",
	}

	// Test create
	err := repo.CreateDependency(ctx, dependency)
	assert.NoError(t, err)

	// Verify dependency was created
	var createdDependency models.PipelineDependency

	err = db.First(&createdDependency, "id = ?", dependency.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, pipeline1.ID, createdDependency.PipelineID)
	assert.Equal(t, pipeline2.ID, createdDependency.DependsOnPipelineID)
	assert.Equal(t, "success", createdDependency.Condition)
}

func TestPipelineRepository_GetDependenciesByPipelineID(t *testing.T) {
	t.Parallel()

	db := setupPipelineTestDB(t)
	repo := NewPipelineRepository(db, nil)
	ctx := context.Background()

	// Create test data
	user := createPipelineTestUser(t, db, "test@example.com")
	project := createPipelineTestProject(t, db, "Test Project", &user.ID)
	pipeline1 := createTestPipeline(t, db, "pipeline1", &project.ID)
	pipeline2 := createTestPipeline(t, db, "pipeline2", &project.ID)
	pipeline3 := createTestPipeline(t, db, "pipeline3", &project.ID)

	// Create dependencies
	dependency1 := &models.PipelineDependency{
		ID:                  uuid.Must(uuid.NewV7()),
		PipelineID:          pipeline1.ID,
		DependsOnPipelineID: pipeline2.ID,
		Condition:           "success",
	}
	dependency2 := &models.PipelineDependency{
		ID:                  uuid.Must(uuid.NewV7()),
		PipelineID:          pipeline1.ID,
		DependsOnPipelineID: pipeline3.ID,
		Condition:           "completed",
	}

	db.Create(dependency1)
	db.Create(dependency2)

	// Get dependencies
	dependencies, err := repo.GetDependenciesByPipelineID(
		ctx,
		pipeline1.ID.String(),
	)
	assert.NoError(t, err)
	assert.Len(t, dependencies, 2)

	// Verify depends-on pipelines are loaded
	for _, dep := range dependencies {
		assert.NotNil(t, dep.DependsOn)
	}
}

func TestPipelineRepository_DeleteDependency(t *testing.T) {
	t.Parallel()

	db := setupPipelineTestDB(t)
	repo := NewPipelineRepository(db, nil)
	ctx := context.Background()

	// Create test data
	user := createPipelineTestUser(t, db, "test@example.com")
	project := createPipelineTestProject(t, db, "Test Project", &user.ID)
	pipeline1 := createTestPipeline(t, db, "pipeline1", &project.ID)
	pipeline2 := createTestPipeline(t, db, "pipeline2", &project.ID)

	// Create dependency
	dependency := &models.PipelineDependency{
		ID:                  uuid.Must(uuid.NewV7()),
		PipelineID:          pipeline1.ID,
		DependsOnPipelineID: pipeline2.ID,
	}
	db.Create(dependency)

	// Delete dependency
	err := repo.DeleteDependency(
		ctx,
		pipeline1.ID.String(),
		pipeline2.ID.String(),
	)
	assert.NoError(t, err)

	// Verify deletion
	var deletedDependency models.PipelineDependency

	err = db.First(
		&deletedDependency,
		"pipeline_id = ? AND depends_on_pipeline_id = ?",
		pipeline1.ID,
		pipeline2.ID,
	).Error
	assert.Error(t, err)
	assert.Equal(t, gorm.ErrRecordNotFound, err)
}

func TestPipelineRepository_GetDependencyGraph(t *testing.T) {
	t.Parallel()

	db := setupPipelineTestDB(t)
	repo := NewPipelineRepository(db, nil)
	ctx := context.Background()

	// Create test data
	user := createPipelineTestUser(t, db, "test@example.com")
	project := createPipelineTestProject(t, db, "Test Project", &user.ID)
	pipeline1 := createTestPipeline(t, db, "pipeline1", &project.ID)
	pipeline2 := createTestPipeline(t, db, "pipeline2", &project.ID)
	pipeline3 := createTestPipeline(t, db, "pipeline3", &project.ID)

	// Create dependencies: pipeline1 -> pipeline2, pipeline1 -> pipeline3
	dependency1 := &models.PipelineDependency{
		ID:                  uuid.Must(uuid.NewV7()),
		PipelineID:          pipeline1.ID,
		DependsOnPipelineID: pipeline2.ID,
	}
	dependency2 := &models.PipelineDependency{
		ID:                  uuid.Must(uuid.NewV7()),
		PipelineID:          pipeline1.ID,
		DependsOnPipelineID: pipeline3.ID,
	}

	db.Create(dependency1)
	db.Create(dependency2)

	// Get dependency graph
	graph, err := repo.GetDependencyGraph(ctx, project.ID.String())
	assert.NoError(t, err)

	// Verify graph structure
	assert.Len(t, graph, 1) // Only pipeline1 has dependencies
	assert.Contains(t, graph, pipeline1.ID.String())
	assert.Len(t, graph[pipeline1.ID.String()], 2)
	assert.Contains(t, graph[pipeline1.ID.String()], pipeline2.ID.String())
	assert.Contains(t, graph[pipeline1.ID.String()], pipeline3.ID.String())
}
