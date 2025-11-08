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

// Package database provides migration functionality for the YAWN application.
package database

import (
	"fmt"

	"github.com/ditwrd/yawn/api/internal/domain/models"
	"gorm.io/gorm"
)

// Migrate runs the auto-migration for all models with proper constraints.
// It creates tables, indexes, foreign keys, and check constraints.
func Migrate(db *gorm.DB) error {
	// Enable foreign key constraints for SQLite
	if err := enableForeignKeysForSQLite(db); err != nil {
		return fmt.Errorf("failed to enable foreign keys for SQLite: %w", err)
	}

	// Run auto-migration for all models
	if err := db.AutoMigrate(
		&models.User{},
		&models.Project{},
		&models.Asset{},
		&models.Repository{},
		&models.Pipeline{},
		&models.ProjectUser{},
		&models.AssetPipeline{},
	); err != nil {
		return fmt.Errorf("failed to run auto-migration: %w", err)
	}

	// Create additional constraints and indexes that aren't handled by AutoMigrate
	if err := createAdditionalConstraints(db); err != nil {
		return fmt.Errorf("failed to create additional constraints: %w", err)
	}

	return nil
}

// enableForeignKeysForSQLite enables foreign key constraints for SQLite databases.
func enableForeignKeysForSQLite(db *gorm.DB) error {
	var sqlDBDialectorName string
	if dialector, ok := db.Dialector.(interface{ Name() string }); ok {
		sqlDBDialectorName = dialector.Name()
	}

	// Only enable foreign keys for SQLite
	if sqlDBDialectorName == "sqlite" {
		return db.Exec("PRAGMA foreign_keys = ON").Error
	}

	return nil
}

// createAdditionalConstraints creates any additional constraints that need manual setup.
func createAdditionalConstraints(db *gorm.DB) error {
	// Create composite unique index for ProjectUser lookup
	if err := db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_project_users_project_user
		ON project_users (project_id, user_id)
	`).Error; err != nil {
		return fmt.Errorf("failed to create unique index for project_users: %w", err)
	}

	// Create index for AssetPipeline ordering
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_asset_pipelines_pipeline_order
		ON asset_pipelines (pipeline_id, "order")
	`).Error; err != nil {
		return fmt.Errorf("failed to create composite index for asset_pipelines: %w", err)
	}

	return nil
}

// DropAllTables drops all tables. Useful for testing and development.
// WARNING: This will delete all data.
func DropAllTables(db *gorm.DB) error {
	// Get all table names
	var tables []string
	if err := db.Raw(`
		SELECT name FROM sqlite_master
		WHERE type='table' AND name NOT LIKE 'sqlite_%' AND name NOT LIKE 'gomigrate_%'
	`).Scan(&tables).Error; err != nil {
		return fmt.Errorf("failed to get table names: %w", err)
	}

	// Drop tables in reverse order to respect foreign key constraints
	for _, table := range tables {
		if err := db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", table)).Error; err != nil {
			return fmt.Errorf("failed to drop table %s: %w", table, err)
		}
	}

	return nil
}
