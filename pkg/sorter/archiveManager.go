package sorter

import (
	"compress/gzip"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/stefanistkuhl/i-would-never-extend-exercises-itsi-y4-ex2/pkg/config"
	"github.com/stefanistkuhl/i-would-never-extend-exercises-itsi-y4-ex2/pkg/db"
	"github.com/stefanistkuhl/i-would-never-extend-exercises-itsi-y4-ex2/pkg/db/sqlc"
	"github.com/stefanistkuhl/i-would-never-extend-exercises-itsi-y4-ex2/pkg/logger"
)

type ArchiveManager struct {
	logger logger.Logger
	cfg    config.Config
}

func NewArchiveManager(cfg config.Config, logger logger.Logger) *ArchiveManager {
	return &ArchiveManager{logger: logger, cfg: cfg}
}
func (am *ArchiveManager) InitialCheck() error {
	if am.cfg.LogLevel == "info" {
		am.logger.Info("Checking for pending archive tasks")
	}

	s, err := db.InitIfNeeded()
	if err != nil {
		am.logger.Error("Failed to get database queries", "error", err)
		return err
	}
	rows, queryErr := s.Queries.GetCapturesForArchive(context.Background(), fmt.Sprintf("-%d days", am.cfg.ArchiveDays))
	if queryErr != nil {
		am.logger.Error("Failed to query captures for archive", "error", queryErr)
	}
	if am.cfg.LogLevel == "info" {
		am.logger.Info("Found captures to archive", "count", len(rows))
	}
	for _, row := range rows {
		if am.cfg.LogLevel == "info" {
			am.logger.Info("Processing capture", "id", row.ID, "path", row.FilePath)
		}
		if am.cfg.CompressionEnabled {
			compErr := am.compressFile(int(row.ID), row.FilePath)
			if compErr != nil {
				am.logger.Error("Failed to compress file", "error", compErr)
				continue
			}
			if !strings.HasSuffix(row.FilePath, ".gz") {
				row.FilePath = row.FilePath + ".gz"
			}
		}
		archErr := am.archiveFile(int(row.ID), row.FilePath)
		if archErr != nil {
			am.logger.Error("Failed to archive file", "error", archErr)
		}
	}

	if cleanupErr := am.cleanupOldArchivedFiles(); cleanupErr != nil {
		am.logger.Error("Failed to cleanup old archived files", "error", cleanupErr)
	}

	if cleanupDirErr := am.cleanupEmptyDirectories(); cleanupDirErr != nil {
		am.logger.Error("Failed to cleanup empty directories", "error", cleanupDirErr)
	}

	return nil
}

func (am *ArchiveManager) StartPeriodicCheck() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		am.logger.Info("Running periodic archive check")

		s, err := db.InitIfNeeded()
		if err != nil {
			am.logger.Error("Failed to get database queries", "error", err)
		}
		rows, queryErr := s.Queries.GetCapturesForArchive(context.Background(), fmt.Sprintf("-%d days", am.cfg.ArchiveDays))
		if queryErr != nil {
			am.logger.Error("Failed to query captures for archive", "error", queryErr)
		}
		if am.cfg.LogLevel == "info" {
			am.logger.Info("Found captures to archive", "count", len(rows))
		}
		for _, row := range rows {
			if am.cfg.LogLevel == "info" {
				am.logger.Info("Processing capture", "id", row.ID, "path", row.FilePath)
			}
			if am.cfg.CompressionEnabled {
				compErr := am.compressFile(int(row.ID), row.FilePath)
				if compErr != nil {
					am.logger.Error("Failed to compress file", "error", compErr)
					continue
				}
				if !strings.HasSuffix(row.FilePath, ".gz") {
					row.FilePath = row.FilePath + ".gz"
				}
			}
			archErr := am.archiveFile(int(row.ID), row.FilePath)
			if archErr != nil {
				am.logger.Error("Failed to archive file", "error", archErr)
			}
		}

		if cleanupErr := am.cleanupOldArchivedFiles(); cleanupErr != nil {
			am.logger.Error("Failed to cleanup old archived files", "error", cleanupErr)
		}

		if cleanupDirErr := am.cleanupEmptyDirectories(); cleanupDirErr != nil {
			am.logger.Error("Failed to cleanup empty directories", "error", cleanupDirErr)
		}

	}
}

func (am *ArchiveManager) archiveFile(id int, filePath string) error {
	if am.cfg.LogLevel == "info" {
		am.logger.Info("Archiving file", "path", filePath, "id", id)
	}

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", filePath)
	}

	s, err := db.InitIfNeeded()
	if err != nil {
		return fmt.Errorf("failed to get database queries: %w", err)
	}
	err = s.Queries.MarkCaptureAsArchived(context.Background(), int64(id))
	if err != nil {
		return fmt.Errorf("failed to mark capture as archived: %w", err)
	}

	relPath, relErr := filepath.Rel(am.cfg.OrganizedDir, filePath)
	if relErr != nil {
		return fmt.Errorf("failed to calculate relative path: %w", relErr)
	}

	targetPath := filepath.Join(am.cfg.ArchiveDir, relPath)

	targetDir := filepath.Dir(targetPath)
	if err := os.MkdirAll(targetDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create archive directory: %w", err)
	}

	renameError := os.Rename(filePath, targetPath)
	if renameError != nil {
		return fmt.Errorf("failed to rename file: %w", renameError)
	}

	updateErr := s.Queries.UpdateFilePath(context.Background(), sqlc.UpdateFilePathParams{
		FilePath: targetPath,
		ID:       int64(id),
	})
	if updateErr != nil {
		am.logger.Error("Failed to update file path in database after archiving", "error", updateErr)
	}

	return nil
}

func (am *ArchiveManager) compressFile(id int, filePath string) error {
	if am.cfg.LogLevel == "info" {
		am.logger.Info("Compressing file", "path", filePath, "id", id)
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", filePath)
	}

	if strings.HasSuffix(filePath, ".gz") {
		if am.cfg.LogLevel == "info" {
			am.logger.Info("File already compressed", "path", filePath)
		}
		return nil
	}

	s, err := db.InitIfNeeded()
	if err != nil {
		return fmt.Errorf("failed to get database queries: %w", err)
	}

	fr, openError := os.ReadFile(filePath)
	if openError != nil {
		return fmt.Errorf("failed to open file: %w", openError)
	}

	compressedPath := filePath + ".gz"
	fa, createError := os.Create(compressedPath)
	if createError != nil {
		return fmt.Errorf("failed to create compressed file: %w", createError)
	}

	w := gzip.NewWriter(fa)
	_, writeErr := w.Write(fr)
	closeErr := w.Close()
	if writeErr != nil {
		fa.Close()
		os.Remove(compressedPath)
		return fmt.Errorf("failed to write to compressed file: %w", writeErr)
	}
	if closeErr != nil {
		fa.Close()
		os.Remove(compressedPath)
		return fmt.Errorf("failed to close gzip writer: %w", closeErr)
	}
	if err := fa.Close(); err != nil {
		os.Remove(compressedPath)
		return fmt.Errorf("failed to close compressed file: %w", err)
	}

	if removeErr := os.Remove(filePath); removeErr != nil {
		am.logger.Error("Failed to remove original file after compression", "path", filePath, "error", removeErr)
	}

	updateErr := s.Queries.UpdateFilePath(context.Background(), sqlc.UpdateFilePathParams{
		FilePath: compressedPath,
		ID:       int64(id),
	})
	if updateErr != nil {
		am.logger.Error("Failed to update file path in database", "error", updateErr)
	}

	err = s.Queries.MarkCaptureAsCompressed(context.Background(), int64(id))
	if err != nil {
		return fmt.Errorf("failed to mark capture as compressed: %w", err)
	}

	return nil
}

func (am *ArchiveManager) cleanupOldArchivedFiles() error {
	if am.cfg.LogLevel == "info" {
		am.logger.Info("Starting cleanup of old archived files", "max_retention_days", am.cfg.MaxRetentionDays)
	}

	s, err := db.InitIfNeeded()
	if err != nil {
		am.logger.Error("Failed to get database queries", "error", err)
		return err
	}

	rows, queryErr := s.Queries.GetOldArchivedCaptures(context.Background(), fmt.Sprintf("-%d days", am.cfg.MaxRetentionDays))
	if queryErr != nil {
		am.logger.Error("Failed to query old archived captures", "error", queryErr)
		return queryErr
	}

	if am.cfg.LogLevel == "info" {
		am.logger.Info("Found old archived captures to delete", "count", len(rows))
	}

	deletedCount := 0
	for _, row := range rows {
		if _, err := os.Stat(row.FilePath); err == nil {
			if err := os.Remove(row.FilePath); err != nil {
				am.logger.Error("Failed to delete old archived file", "path", row.FilePath, "error", err)
				continue
			}
			if am.cfg.LogLevel == "info" {
				am.logger.Info("Deleted old archived file", "path", row.FilePath, "id", row.ID)
			}
		} else if !os.IsNotExist(err) {
			am.logger.Error("Failed to stat file before deletion", "path", row.FilePath, "error", err)
			continue
		}

		if err := s.Queries.DeleteCapture(context.Background(), row.ID); err != nil {
			am.logger.Error("Failed to delete capture from database", "id", row.ID, "error", err)
			continue
		}

		deletedCount++
	}

	if am.cfg.LogLevel == "info" {
		am.logger.Info("Completed cleanup of old archived files", "deleted_count", deletedCount)
	}

	return nil
}

func (am *ArchiveManager) cleanupEmptyDirectories() error {
	if am.cfg.LogLevel == "info" {
		am.logger.Info("Starting cleanup of empty directories", "organized_dir", am.cfg.OrganizedDir)
	}

	cleanedCount := 0
	err := filepath.Walk(am.cfg.OrganizedDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if path == am.cfg.OrganizedDir {
			return nil
		}

		if !info.IsDir() {
			return nil
		}

		entries, err := os.ReadDir(path)
		if err != nil {
			return nil
		}

		if len(entries) == 0 {
			if err := os.Remove(path); err != nil {
				am.logger.Error("Failed to remove empty directory", "path", path, "error", err)
				return nil
			}
			cleanedCount++
			if am.cfg.LogLevel == "info" {
				am.logger.Info("Removed empty directory", "path", path)
			}
		}

		return nil
	})

	cleanedMore := true
	for cleanedMore {
		cleanedMore = false
		err := filepath.Walk(am.cfg.OrganizedDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}

			if path == am.cfg.OrganizedDir {
				return nil
			}

			if !info.IsDir() {
				return nil
			}

			entries, err := os.ReadDir(path)
			if err != nil {
				return nil
			}

			if len(entries) == 0 {
				if err := os.Remove(path); err != nil {
					return nil
				}
				cleanedCount++
				cleanedMore = true
				if am.cfg.LogLevel == "info" {
					am.logger.Info("Removed empty directory", "path", path)
				}
			}

			return nil
		})
		if err != nil {
			am.logger.Error("Error during empty directory cleanup", "error", err)
		}
	}

	if am.cfg.LogLevel == "info" {
		am.logger.Info("Completed cleanup of empty directories", "cleaned_count", cleanedCount)
	}

	return err
}
