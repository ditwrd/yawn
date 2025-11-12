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

// Package services provides business logic layer implementations for domain
// entities.
//
// This package contains service interfaces and implementations that encapsulate
// business rules, validation, and orchestration of repository operations.
// All services are context-aware and include proper error handling and logging.
package services

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/gofrs/uuid"
	"github.com/rs/zerolog"

	"github.com/ditwrd/yawn/api/internal/domain/models"
	"github.com/ditwrd/yawn/api/internal/domain/repositories"
)

// GitOpsService defines the interface for Git operations and pipeline
// synchronization.
//
// Provides methods for repository cloning, pipeline parsing, synchronization,
// and webhook handling for GitOps workflows.
type GitOpsService interface {
	// CloneRepository clones a Git repository to a local directory
	CloneRepository(
		ctx context.Context,
		repoURL, localPath string,
	) (*git.Repository, error)

	// ParsePipelinesFromGit discovers and parses pipeline definitions from a Git
	// repository
	ParsePipelinesFromGit(
		ctx context.Context,
		repo *git.Repository,
		projectID string,
	) ([]*PipelineDefinition, error)

	// SyncRepository synchronizes a repository and updates pipelines
	SyncRepository(ctx context.Context, repositoryID string) (*SyncResult, error)

	// HandleWebhook processes Git webhook events to trigger synchronization
	HandleWebhook(ctx context.Context, webhookPayload *WebhookPayload) error

	// GetLatestCommit retrieves the latest commit information for a repository
	GetLatestCommit(
		ctx context.Context,
		repoURL, branch string,
	) (*CommitInfo, error)

	// ValidateGitRepository validates that a Git repository is accessible and
	// valid
	ValidateGitRepository(ctx context.Context, repoURL string) error
}

// PipelineDefinition represents a pipeline definition parsed from Git.
type PipelineDefinition struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Path        string         `json:"path"`
	Content     string         `json:"content"`
	Hash        string         `json:"hash"`
	Metadata    map[string]any `json:"metadata"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// SyncResult represents the result of a repository synchronization.
type SyncResult struct {
	Success    bool          `json:"success"`
	CommitHash string        `json:"commit_hash"`
	Message    string        `json:"message"`
	Pipelines  []string      `json:"pipelines"`
	Errors     []string      `json:"errors,omitempty"`
	Duration   time.Duration `json:"duration"`
	SyncedAt   time.Time     `json:"synced_at"`
	Changes    SyncChanges   `json:"changes"`
}

// SyncChanges represents changes detected during synchronization.
type SyncChanges struct {
	Added    []string `json:"added"`
	Modified []string `json:"modified"`
	Deleted  []string `json:"deleted"`
}

// WebhookPayload represents a Git webhook payload.
type WebhookPayload struct {
	EventType string         `json:"event_type"`
	RepoURL   string         `json:"repo_url"`
	Branch    string         `json:"branch"`
	Commit    string         `json:"commit"`
	Timestamp time.Time      `json:"timestamp"`
	Payload   map[string]any `json:"payload"`
}

// CommitInfo represents Git commit information.
type CommitInfo struct {
	Hash      string    `json:"hash"`
	Message   string    `json:"message"`
	Author    string    `json:"author"`
	Email     string    `json:"email"`
	Timestamp time.Time `json:"timestamp"`
	Branch    string    `json:"branch"`
}

// gitOpsService implements the GitOpsService interface.
type gitOpsService struct {
	repoRepo     repositories.RepositoryRepository
	pipelineRepo repositories.PipelineRepository
	projectRepo  repositories.ProjectRepository
	logger       *zerolog.Logger
	tempDir      string
	syncTimeout  time.Duration
}

// NewGitOpsService creates a new instance of GitOpsService
//
// Parameters:
//   - repoRepo: Repository repository for data operations
//   - pipelineRepo: Pipeline repository for pipeline operations
//   - projectRepo: Project repository for project operations
//   - logger: Logger for structured logging
//
// Returns:
//   - GitOpsService: An instance of the GitOps service
func NewGitOpsService(
	repoRepo repositories.RepositoryRepository,
	pipelineRepo repositories.PipelineRepository,
	projectRepo repositories.ProjectRepository,
	logger *zerolog.Logger,
) GitOpsService {
	// Create temporary directory for Git operations
	tempDir := filepath.Join(os.TempDir(), "yawn-gitops")

	err := os.MkdirAll(tempDir, 0o755)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to create GitOps temporary directory")
	}

	return &gitOpsService{
		repoRepo:     repoRepo,
		pipelineRepo: pipelineRepo,
		projectRepo:  projectRepo,
		logger:       logger,
		tempDir:      tempDir,
		syncTimeout:  10 * time.Minute, // Default timeout for sync operations
	}
}

// CloneRepository clones a Git repository to a local directory.
func (s *gitOpsService) CloneRepository(
	ctx context.Context,
	repoURL, localPath string,
) (*git.Repository, error) {
	s.logger.Info().
		Str("repo_url", repoURL).
		Str("local_path", localPath).
		Msg("Cloning Git repository")

	// Check if repository already exists and remove it
	if _, err := os.Stat(localPath); err == nil {
		s.logger.Info().Str("path", localPath).Msg("Removing existing repository")
		os.RemoveAll(localPath)
	}

	// Clone the repository
	repo, err := git.PlainCloneContext(ctx, localPath, false, &git.CloneOptions{
		URL:      repoURL,
		Depth:    1, // Shallow clone for performance
		Progress: os.Stdout,
	})
	if err != nil {
		s.logger.Error().
			Err(err).
			Str("repo_url", repoURL).
			Msg("Failed to clone repository")

		return nil, fmt.Errorf("failed to clone repository %s: %w", repoURL, err)
	}

	s.logger.Info().
		Str("repo_url", repoURL).
		Str("local_path", localPath).
		Msg("Repository cloned successfully")

	return repo, nil
}

// ParsePipelinesFromGit discovers and parses pipeline definitions from a Git
// repository.
func (s *gitOpsService) ParsePipelinesFromGit(
	ctx context.Context,
	repo *git.Repository,
	projectID string,
) ([]*PipelineDefinition, error) {
	s.logger.Info().
		Str("project_id", projectID).
		Msg("Parsing pipelines from Git repository")

	var pipelines []*PipelineDefinition

	// Get the worktree
	worktree, err := repo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("failed to get worktree: %w", err)
	}

	// Define pipeline file patterns to search for
	pipelinePatterns := []string{
		"**/*.yml",
		"**/*.yaml",
		"**/pipeline.yml",
		"**/pipeline.yaml",
		"**/.yawn/*.yml",
		"**/.yawn/*.yaml",
	}

	// Search for pipeline files
	for _, pattern := range pipelinePatterns {
		files, err := s.findPipelineFiles(worktree, pattern)
		if err != nil {
			s.logger.Warn().
				Err(err).
				Str("pattern", pattern).
				Msg("Failed to search for pipeline files")

			continue
		}

		for _, file := range files {
			pipeline, err := s.parsePipelineFile(worktree, file)
			if err != nil {
				s.logger.Warn().
					Err(err).
					Str("file", file).
					Msg("Failed to parse pipeline file")

				continue
			}

			if pipeline != nil {
				pipelines = append(pipelines, pipeline)
			}
		}
	}

	s.logger.Info().
		Str("project_id", projectID).
		Int("pipeline_count", len(pipelines)).
		Msg("Successfully parsed pipelines from Git")

	return pipelines, nil
}

// SyncRepository synchronizes a repository and updates pipelines.
func (s *gitOpsService) SyncRepository(
	ctx context.Context,
	repositoryID string,
) (*SyncResult, error) {
	startTime := time.Now()

	s.logger.Info().
		Str("repository_id", repositoryID).
		Msg("Starting repository synchronization")

	result := &SyncResult{
		SyncedAt: startTime,
		Changes:  SyncChanges{},
	}

	// Get repository information
	repo, err := s.repoRepo.GetByID(ctx, repositoryID)
	if err != nil {
		result.Success = false
		result.Message = fmt.Sprintf("Failed to get repository: %v", err)
		result.Duration = time.Since(startTime)

		return result, err
	}

	// Get project information
	project, err := s.projectRepo.GetByID(repo.ProjectID.String())
	if err != nil {
		result.Success = false
		result.Message = fmt.Sprintf("Failed to get project: %v", err)
		result.Duration = time.Since(startTime)

		return result, err
	}

	// Create temporary directory for this sync
	tempPath := filepath.Join(
		s.tempDir,
		fmt.Sprintf(
			"sync-%s-%s",
			repo.ID.String(),
			time.Now().Format("20060102-150405"),
		),
	)
	defer os.RemoveAll(tempPath)

	// Clone repository
	gitRepo, err := s.CloneRepository(ctx, repo.URL, tempPath)
	if err != nil {
		result.Success = false
		result.Message = fmt.Sprintf("Failed to clone repository: %v", err)
		result.Duration = time.Since(startTime)

		return result, err
	}

	// Get latest commit hash
	head, err := gitRepo.Head()
	if err != nil {
		result.Success = false
		result.Message = fmt.Sprintf("Failed to get repository head: %v", err)
		result.Duration = time.Since(startTime)

		return result, err
	}

	currentCommit := head.Hash().String()

	// Check if repository is already up to date
	if repo.LatestCommit == currentCommit {
		result.Success = true
		result.Message = "Repository is already up to date"
		result.CommitHash = currentCommit
		result.Duration = time.Since(startTime)

		return result, nil
	}

	// Parse pipelines from Git
	pipelineDefs, err := s.ParsePipelinesFromGit(
		ctx,
		gitRepo,
		project.ID.String(),
	)
	if err != nil {
		result.Success = false
		result.Message = fmt.Sprintf("Failed to parse pipelines: %v", err)
		result.Duration = time.Since(startTime)

		return result, err
	}

	// Update repository sync status
	repo.LatestCommit = currentCommit

	repo.SyncStatus = models.RepositoryStatusSuccess
	if err := s.repoRepo.Update(ctx, repo); err != nil {
		s.logger.Error().
			Err(err).
			Str("repository_id", repositoryID).
			Msg("Failed to update repository sync status")
		// Don't fail the sync if we can't update the status
	}

	// Process pipeline definitions
	var pipelineNames []string
	for _, pipelineDef := range pipelineDefs {
		pipelineNames = append(pipelineNames, pipelineDef.Name)

		// Check if pipeline already exists
		existing, err := s.pipelineRepo.ExistsByName(
			ctx,
			project.ID.String(),
			pipelineDef.Name,
		)
		if err != nil {
			s.logger.Error().
				Err(err).
				Str("pipeline_name", pipelineDef.Name).
				Msg("Failed to check if pipeline exists")
			result.Errors = append(
				result.Errors,
				fmt.Sprintf("Pipeline %s: failed to check existence", pipelineDef.Name),
			)

			continue
		}

		if existing {
			// Update existing pipeline
			err = s.updatePipelineFromDefinition(
				ctx,
				project.ID.String(),
				pipelineDef,
			)
			if err != nil {
				s.logger.Error().
					Err(err).
					Str("pipeline_name", pipelineDef.Name).
					Msg("Failed to update pipeline")
				result.Errors = append(
					result.Errors,
					fmt.Sprintf("Pipeline %s: failed to update", pipelineDef.Name),
				)
			} else {
				result.Changes.Modified = append(result.Changes.Modified, pipelineDef.Name)
			}
		} else {
			// Create new pipeline
			err = s.createPipelineFromDefinition(ctx, project.ID.String(), pipelineDef)
			if err != nil {
				s.logger.Error().
					Err(err).
					Str("pipeline_name", pipelineDef.Name).
					Msg("Failed to create pipeline")
				result.Errors = append(result.Errors, fmt.Sprintf("Pipeline %s: failed to create", pipelineDef.Name))
			} else {
				result.Changes.Added = append(result.Changes.Added, pipelineDef.Name)
			}
		}
	}

	// Prepare result
	result.Success = true
	result.CommitHash = currentCommit
	result.Message = "Repository synchronized successfully"
	result.Pipelines = pipelineNames
	result.Duration = time.Since(startTime)

	s.logger.Info().
		Str("repository_id", repositoryID).
		Str("commit_hash", currentCommit).
		Int("pipelines_found", len(pipelineDefs)).
		Int("pipelines_added", len(result.Changes.Added)).
		Int("pipelines_modified", len(result.Changes.Modified)).
		Dur("duration", result.Duration).
		Msg("Repository synchronization completed")

	return result, nil
}

// HandleWebhook processes Git webhook events to trigger synchronization.
func (s *gitOpsService) HandleWebhook(
	ctx context.Context,
	webhookPayload *WebhookPayload,
) error {
	s.logger.Info().
		Str("event_type", webhookPayload.EventType).
		Str("repo_url", webhookPayload.RepoURL).
		Str("branch", webhookPayload.Branch).
		Str("commit", webhookPayload.Commit).
		Msg("Processing Git webhook")

	// Find repository by URL
	repos, err := s.repoRepo.ListByURL(ctx, webhookPayload.RepoURL)
	if err != nil {
		return fmt.Errorf(
			"failed to find repositories for URL %s: %w",
			webhookPayload.RepoURL,
			err,
		)
	}

	if len(repos) == 0 {
		s.logger.Warn().
			Str("repo_url", webhookPayload.RepoURL).
			Msg("No repositories found for webhook URL")

		return fmt.Errorf(
			"no repositories found for URL %s",
			webhookPayload.RepoURL,
		)
	}

	// Trigger synchronization for all matching repositories
	for _, repo := range repos {
		// Check if the webhook is for the correct branch
		if repo.Branch != webhookPayload.Branch {
			s.logger.Info().
				Str("repository_id", repo.ID.String()).
				Str("webhook_branch", webhookPayload.Branch).
				Str("repo_branch", repo.Branch).
				Msg("Skipping sync due to branch mismatch")

			continue
		}

		s.logger.Info().
			Str("repository_id", repo.ID.String()).
			Msg("Triggering synchronization from webhook")

		_, err := s.SyncRepository(ctx, repo.ID.String())
		if err != nil {
			s.logger.Error().
				Err(err).
				Str("repository_id", repo.ID.String()).
				Msg("Failed to synchronize repository from webhook")

			return fmt.Errorf(
				"failed to sync repository %s: %w",
				repo.ID.String(),
				err,
			)
		}
	}

	s.logger.Info().
		Str("repo_url", webhookPayload.RepoURL).
		Int("repositories_synced", len(repos)).
		Msg("Webhook processing completed")

	return nil
}

// GetLatestCommit retrieves the latest commit information for a repository.
func (s *gitOpsService) GetLatestCommit(
	ctx context.Context,
	repoURL, branch string,
) (*CommitInfo, error) {
	s.logger.Info().
		Str("repo_url", repoURL).
		Str("branch", branch).
		Msg("Getting latest commit information")

	// Create temporary directory for clone
	tempPath := filepath.Join(
		s.tempDir,
		fmt.Sprintf(
			"commit-%s-%s",
			uuid.Must(uuid.NewV7()).String(),
			time.Now().Format("20060102-150405"),
		),
	)
	defer os.RemoveAll(tempPath)

	// Clone repository
	repo, err := git.PlainCloneContext(ctx, tempPath, false, &git.CloneOptions{
		URL:   repoURL,
		Depth: 1,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to clone repository: %w", err)
	}

	// Get head reference
	head, err := repo.Head()
	if err != nil {
		return nil, fmt.Errorf("failed to get repository head: %w", err)
	}

	// Get commit object
	commit, err := repo.CommitObject(head.Hash())
	if err != nil {
		return nil, fmt.Errorf("failed to get commit object: %w", err)
	}

	commitInfo := &CommitInfo{
		Hash:      commit.Hash.String(),
		Message:   commit.Message,
		Author:    commit.Author.Name,
		Email:     commit.Author.Email,
		Timestamp: commit.Author.When,
		Branch:    branch,
	}

	s.logger.Info().
		Str("repo_url", repoURL).
		Str("commit_hash", commitInfo.Hash).
		Str("author", commitInfo.Author).
		Msg("Retrieved latest commit information")

	return commitInfo, nil
}

// ValidateGitRepository validates that a Git repository is accessible and
// valid.
func (s *gitOpsService) ValidateGitRepository(
	ctx context.Context,
	repoURL string,
) error {
	s.logger.Info().
		Str("repo_url", repoURL).
		Msg("Validating Git repository")

	// Create temporary directory for validation
	tempPath := filepath.Join(
		s.tempDir,
		fmt.Sprintf(
			"validate-%s-%s",
			uuid.Must(uuid.NewV7()).String(),
			time.Now().Format("20060102-150405"),
		),
	)
	defer os.RemoveAll(tempPath)

	// Try to clone repository (shallow clone for validation)
	repo, err := git.PlainCloneContext(ctx, tempPath, false, &git.CloneOptions{
		URL:   repoURL,
		Depth: 1,
	})
	if err != nil {
		s.logger.Error().
			Err(err).
			Str("repo_url", repoURL).
			Msg("Repository validation failed")

		return fmt.Errorf("repository validation failed: %w", err)
	}

	// Verify it's a valid Git repository
	_, err = repo.Head()
	if err != nil {
		return fmt.Errorf("invalid repository: %w", err)
	}

	s.logger.Info().
		Str("repo_url", repoURL).
		Msg("Repository validation successful")

	return nil
}

// Helper methods

func (s *gitOpsService) findPipelineFiles(
	worktree *git.Worktree,
	pattern string,
) ([]string, error) {
	var files []string

	// Get all files in the repository
	fileIter, err := worktree.Filesystem.ReadDir(".")
	if err != nil {
		return nil, fmt.Errorf("failed to read filesystem: %w", err)
	}

	// Simple pattern matching (can be enhanced with proper glob patterns)
	for _, file := range fileIter {
		if s.matchesPipelinePattern(file.Name(), pattern) {
			files = append(files, file.Name())
		}
	}

	return files, nil
}

func (s *gitOpsService) matchesPipelinePattern(filename, pattern string) bool {
	// Simple pattern matching for pipeline files
	pipelineKeywords := []string{"pipeline", ".yawn", "workflow", "ci", "cd"}

	for _, keyword := range pipelineKeywords {
		if strings.Contains(strings.ToLower(filename), keyword) {
			return true
		}
	}

	// Check file extensions
	ext := filepath.Ext(filename)

	return ext == ".yml" || ext == ".yaml"
}

func (s *gitOpsService) parsePipelineFile(
	worktree *git.Worktree,
	filePath string,
) (*PipelineDefinition, error) {
	// Read file content
	content, err := worktree.Filesystem.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer content.Close()

	// For now, return a basic pipeline definition
	// In a real implementation, you would parse the YAML/YML content
	// and extract pipeline configuration
	return &PipelineDefinition{
		Name:        filepath.Base(filePath),
		Description: "Pipeline from " + filePath,
		Path:        filePath,
		Content:     "", // Would contain actual file content
		Hash:        "", // Would be calculated from content
		Metadata: map[string]any{
			"source": "git",
			"path":   filePath,
		},
		UpdatedAt: time.Now(),
	}, nil
}

func (s *gitOpsService) createPipelineFromDefinition(
	ctx context.Context,
	projectID string,
	def *PipelineDefinition,
) error {
	pipeline := &models.Pipeline{
		ID:          uuid.Must(uuid.NewV7()),
		Name:        def.Name,
		Description: def.Description,
		ProjectID:   uuid.Must(uuid.FromString(projectID)),
		Status:      models.PipelineStatusDraft,
		Config:      def.Content,
		IsEnabled:   false, // Disabled by default until reviewed
	}

	return s.pipelineRepo.Create(ctx, pipeline)
}

func (s *gitOpsService) updatePipelineFromDefinition(
	ctx context.Context,
	projectID string,
	def *PipelineDefinition,
) error {
	// Find existing pipeline
	pipelines, err := s.pipelineRepo.GetByProjectID(ctx, projectID, 100, 0)
	if err != nil {
		return fmt.Errorf("failed to get existing pipelines: %w", err)
	}

	var targetPipeline *models.Pipeline

	for _, pipeline := range pipelines {
		if pipeline.Name == def.Name {
			targetPipeline = pipeline

			break
		}
	}

	if targetPipeline == nil {
		return fmt.Errorf("pipeline %s not found", def.Name)
	}

	// Update pipeline
	targetPipeline.Description = def.Description
	targetPipeline.Config = def.Content

	return s.pipelineRepo.Update(ctx, targetPipeline)
}
