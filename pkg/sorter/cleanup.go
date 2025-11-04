package sorter

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/stefanistkuhl/i-would-never-extend-exercises-itsi-y4-ex2/pkg/config"
	"github.com/stefanistkuhl/i-would-never-extend-exercises-itsi-y4-ex2/pkg/db"
)

func (s *Server) GetCleanupCandidatesHandler(w http.ResponseWriter, r *http.Request) {
	candidates, err := s.getCleanupCandidates()
	if err != nil {
		s.logger.Error("Failed to get cleanup candidates", "error", err)
		jsonResponse(w, http.StatusInternalServerError, StatusRes{Status: "error"})
		return
	}

	jsonResponse(w, http.StatusOK, candidates)
}

func (s *Server) CleanupExecuteHandler(w http.ResponseWriter, r *http.Request) {
	candidates, err := s.getCleanupCandidates()
	if err != nil {
		s.logger.Error("Failed to get cleanup candidates for execution", "error", err)
		jsonResponse(w, http.StatusInternalServerError, StatusRes{Status: "error"})
		return
	}

	if s.isCleanupEmpty(candidates) {
		jsonResponse(w, http.StatusNoContent, CleanupEmptyRes{
			Message: "No cleanup candidates found - nothing to clean up",
			Status:  "empty",
		})
		return
	}

	result := CleanupExecuteRes{
		Errors: []string{},
	}

	cfg := s.GetConfig()

	store, dbErr := db.InitIfNeeded()
	if dbErr != nil {
		s.logger.Error("Failed to initialize database", "error", dbErr)
		jsonResponse(w, http.StatusInternalServerError, StatusRes{Status: "error"})
		return
	}

	for _, file := range candidates.OldArchivedFiles {
		if _, err := os.Stat(file.FilePath); err == nil {
			if err := os.Remove(file.FilePath); err != nil {
				errMsg := fmt.Sprintf("Failed to delete old archived file %s: %v", file.FilePath, err)
				s.logger.Error(errMsg)
				result.Errors = append(result.Errors, errMsg)
				continue
			}
			if cfg.LogLevel == "info" {
				s.logger.Info("Deleted old archived file", "path", file.FilePath, "id", file.ID)
			}
		} else if !os.IsNotExist(err) {
			errMsg := fmt.Sprintf("Failed to stat file before deletion %s: %v", file.FilePath, err)
			s.logger.Error(errMsg)
			result.Errors = append(result.Errors, errMsg)
			continue
		}

		if err := store.DeleteCapture(context.Background(), file.ID); err != nil {
			errMsg := fmt.Sprintf("Failed to delete capture from database (id: %d): %v", file.ID, err)
			s.logger.Error(errMsg)
			result.Errors = append(result.Errors, errMsg)
			continue
		}

		result.DeletedOldArchivedFiles++
	}

	for _, dir := range candidates.EmptyDirectories {
		if err := os.Remove(dir); err != nil {
			errMsg := fmt.Sprintf("Failed to remove empty directory %s: %v", dir, err)
			s.logger.Error(errMsg)
			result.Errors = append(result.Errors, errMsg)
			continue
		}
		if cfg.LogLevel == "info" {
			s.logger.Info("Removed empty directory", "path", dir)
		}
		result.DeletedEmptyDirectories++
	}

	cleanedMore := true
	for cleanedMore {
		cleanedMore = false
		err := filepath.Walk(cfg.OrganizedDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if path == cfg.OrganizedDir {
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
				cleanedMore = true
				result.DeletedEmptyDirectories++
				if cfg.LogLevel == "info" {
					s.logger.Info("Removed empty directory", "path", path)
				}
			}
			return nil
		})
		if err != nil {
			s.logger.Error("Error during empty directory cleanup", "error", err)
		}
	}

	cleanedMore = true
	for cleanedMore {
		cleanedMore = false
		err := filepath.Walk(cfg.ArchiveDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if path == cfg.ArchiveDir {
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
				cleanedMore = true
				result.DeletedEmptyDirectories++
				if cfg.LogLevel == "info" {
					s.logger.Info("Removed empty directory", "path", path)
				}
			}
			return nil
		})
		if err != nil {
			s.logger.Error("Error during empty directory cleanup", "error", err)
		}
	}

	for _, file := range candidates.UntrackedFiles {
		if err := os.Remove(file.FilePath); err != nil {
			errMsg := fmt.Sprintf("Failed to delete untracked file %s: %v", file.FilePath, err)
			s.logger.Error(errMsg)
			result.Errors = append(result.Errors, errMsg)
			continue
		}
		if cfg.LogLevel == "info" {
			s.logger.Info("Deleted untracked file", "path", file.FilePath)
		}
		result.DeletedUntrackedFiles++
	}

	result.Summary = candidates.Summary
	jsonResponse(w, http.StatusOK, result)
}

func (s *Server) getCleanupCandidates() (CleanupCandidatesRes, error) {
	cfg := s.GetConfig()
	result := CleanupCandidatesRes{
		OldArchivedFiles: []OldArchivedFile{},
		EmptyDirectories: []string{},
		UntrackedFiles:   []UntrackedFile{},
	}

	store, err := db.InitIfNeeded()
	if err != nil {
		return result, fmt.Errorf("failed to initialize database: %w", err)
	}

	oldArchivedRows, err := store.GetOldArchivedCaptures(context.Background(), fmt.Sprintf("-%d days", cfg.MaxRetentionDays))
	if err != nil {
		return result, fmt.Errorf("failed to query old archived captures: %w", err)
	}

	for _, row := range oldArchivedRows {
		result.OldArchivedFiles = append(result.OldArchivedFiles, OldArchivedFile{
			ID:       row.ID,
			FilePath: row.FilePath,
		})
	}

	emptyDirs, err := s.findEmptyDirectories(cfg.OrganizedDir)
	if err != nil {
		s.logger.Error("Failed to find empty directories in organized dir", "error", err)
	} else {
		result.EmptyDirectories = append(result.EmptyDirectories, emptyDirs...)
	}

	emptyArchiveDirs, err := s.findEmptyDirectories(cfg.ArchiveDir)
	if err != nil {
		s.logger.Error("Failed to find empty directories in archive dir", "error", err)
	} else {
		result.EmptyDirectories = append(result.EmptyDirectories, emptyArchiveDirs...)
	}

	untrackedFiles, err := s.findUntrackedFiles(cfg, store)
	if err != nil {
		s.logger.Error("Failed to find untracked files", "error", err)
	} else {
		result.UntrackedFiles = untrackedFiles
	}

	totalSize := int64(0)
	for _, file := range result.OldArchivedFiles {
		if info, err := os.Stat(file.FilePath); err == nil {
			totalSize += info.Size()
		}
	}
	for _, file := range result.UntrackedFiles {
		totalSize += file.Size
	}

	result.Summary = CleanupSummary{
		OldArchivedFilesCount: len(result.OldArchivedFiles),
		EmptyDirectoriesCount: len(result.EmptyDirectories),
		UntrackedFilesCount:   len(result.UntrackedFiles),
		TotalSizeToFree:       totalSize,
	}

	return result, nil
}

func (s *Server) findEmptyDirectories(rootPath string) ([]string, error) {
	var emptyDirs []string

	if _, err := os.Stat(rootPath); os.IsNotExist(err) {
		return emptyDirs, nil
	}

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if path == rootPath {
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
			emptyDirs = append(emptyDirs, path)
		}

		return nil
	})

	return emptyDirs, err
}

func (s *Server) findUntrackedFiles(cfg config.Config, store *db.Store) ([]UntrackedFile, error) {
	var untrackedFiles []UntrackedFile

	allCaptures, err := store.GetCaptures(context.Background())
	if err != nil {
		return untrackedFiles, fmt.Errorf("failed to get captures: %w", err)
	}

	trackedPaths := make(map[string]bool)
	for _, capture := range allCaptures {
		trackedPaths[capture.FilePath] = true
	}

	untrackedInOrg, err := s.findUntrackedFilesInDir(cfg.OrganizedDir, trackedPaths)
	if err != nil {
		s.logger.Error("Failed to find untracked files in organized dir", "error", err)
	} else {
		untrackedFiles = append(untrackedFiles, untrackedInOrg...)
	}

	untrackedInArchive, err := s.findUntrackedFilesInDir(cfg.ArchiveDir, trackedPaths)
	if err != nil {
		s.logger.Error("Failed to find untracked files in archive dir", "error", err)
	} else {
		untrackedFiles = append(untrackedFiles, untrackedInArchive...)
	}

	return untrackedFiles, nil
}

func (s *Server) findUntrackedFilesInDir(dirPath string, trackedPaths map[string]bool) ([]UntrackedFile, error) {
	var untrackedFiles []UntrackedFile

	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return untrackedFiles, nil
	}

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() {
			return nil
		}

		if !trackedPaths[path] {
			untrackedFiles = append(untrackedFiles, UntrackedFile{
				FilePath: path,
				Size:     info.Size(),
			})
		}

		return nil
	})

	return untrackedFiles, err
}

// isCleanupEmpty checks if there are any items to clean up
func (s *Server) isCleanupEmpty(candidates CleanupCandidatesRes) bool {
	return len(candidates.OldArchivedFiles) == 0 &&
		len(candidates.EmptyDirectories) == 0 &&
		len(candidates.UntrackedFiles) == 0
}
