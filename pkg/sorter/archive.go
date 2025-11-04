package sorter

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stefanistkuhl/i-would-never-extend-exercises-itsi-y4-ex2/pkg/db"
	"github.com/stefanistkuhl/i-would-never-extend-exercises-itsi-y4-ex2/pkg/db/sqlc"
)

func (s *Server) GetArchiveHandler(w http.ResponseWriter, r *http.Request) {
	store, dbErr := db.InitIfNeeded()
	if dbErr != nil {
		s.logger.Error("Failed to initialize database", "error", dbErr)
		jsonResponse(w, http.StatusInternalServerError, StatusRes{Status: "error"})
		return
	}
	captures, getCapturesErr := store.GetArchviedCaptures(context.Background())
	if getCapturesErr != nil {
		s.logger.Error("Failed to get captures", "error", getCapturesErr)
		jsonResponse(w, http.StatusInternalServerError, StatusRes{Status: "error"})
		return
	}
	caps, marshalErr := json.Marshal(captures)
	if marshalErr != nil {
		s.logger.Error("Failed to marshal captures", "error", marshalErr)
		jsonResponse(w, http.StatusInternalServerError, StatusRes{Status: "error"})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(caps)
}

func (s *Server) ArchiveFileHandler(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	captureID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		s.logger.Error("Invalid capture ID", "error", err, "id", idParam)
		jsonResponse(w, http.StatusBadRequest, StatusRes{Status: "error"})
		return
	}

	store, dbErr := db.InitIfNeeded()
	if dbErr != nil {
		s.logger.Error("Failed to initialize database", "error", dbErr)
		jsonResponse(w, http.StatusInternalServerError, StatusRes{Status: "error"})
		return
	}

	capture, getCaptureErr := store.GetCapture(context.Background(), captureID)
	if getCaptureErr != nil {
		s.logger.Error("Failed to get capture", "error", getCaptureErr, "id", captureID)
		if getCaptureErr == sql.ErrNoRows {
			jsonResponse(w, http.StatusNotFound, StatusRes{Status: "error"})
		} else {
			jsonResponse(w, http.StatusInternalServerError, StatusRes{Status: "error"})
		}
		return
	}

	if capture.Archived.Bool {
		jsonResponse(w, http.StatusBadRequest, StatusRes{Status: "error"})
		return
	}

	err = store.MarkCaptureAsArchived(context.Background(), captureID)
	if err != nil {
		s.logger.Error("Failed to archive capture", "error", err, "id", captureID)
		jsonResponse(w, http.StatusInternalServerError, StatusRes{Status: "error"})
		return
	}
	cfg := s.GetConfig()
	relPath, relErr := filepath.Rel(cfg.OrganizedDir, capture.FilePath)
	if relErr != nil {
		s.logger.Error("Failed to calculate relative path", "error", relErr, "path", capture.FilePath)
		jsonResponse(w, http.StatusInternalServerError, StatusRes{Status: "error"})
		return
	}

	targetPath := filepath.Join(cfg.ArchiveDir, relPath)

	targetDir := filepath.Dir(targetPath)
	if err := os.MkdirAll(targetDir, os.ModePerm); err != nil {
		s.logger.Error("Failed to create archive directory", "error", err, "path", targetDir)
		jsonResponse(w, http.StatusInternalServerError, StatusRes{Status: "error"})
		return
	}

	renameError := os.Rename(capture.FilePath, targetPath)
	if renameError != nil {
		s.logger.Error("Failed to rename file", "error", renameError, "path", capture.FilePath)
		jsonResponse(w, http.StatusInternalServerError, StatusRes{Status: "error"})
		return
	}

	updateErr := store.UpdateFilePath(context.Background(), sqlc.UpdateFilePathParams{
		FilePath: targetPath,
		ID:       captureID,
	})
	if updateErr != nil {
		s.logger.Error("Failed to update file path in database after archiving", "error", updateErr)
		jsonResponse(w, http.StatusInternalServerError, StatusRes{Status: "error"})
		return
	}

	jsonResponse(w, http.StatusOK, StatusRes{Status: "ok"})
}

func (s *Server) ArchiveStatusHandler(w http.ResponseWriter, r *http.Request) {
	store, dbErr := db.InitIfNeeded()
	if dbErr != nil {
		s.logger.Error("Failed to initialize database", "error", dbErr)
		jsonResponse(w, http.StatusInternalServerError, StatusRes{Status: "error"})
		return
	}
	captures, getCapturesErr := store.GetArchiveBrief(context.Background())
	if getCapturesErr != nil {
		s.logger.Error("Failed to get archive status", "error", getCapturesErr)
		jsonResponse(w, http.StatusInternalServerError, StatusRes{Status: "error"})
		return
	}
	numFiles := len(captures)
	size := int64(0)
	archivedCaptures := make([]ArchivedFile, 0, numFiles)
	for _, capture := range captures {
		size += capture.FileSize
		var archivedCapture ArchivedFile
		archivedCapture.ID = capture.ID
		archivedCapture.FilePath = capture.FilePath
		archivedCapture.FileSize = int64(capture.FileSize)
		archivedCapture.CreatedAt = capture.CreatedAt.Time.Format(time.RFC3339)
		archivedCapture.UpdatedAt = capture.UpdatedAt.Time.Format(time.RFC3339)
		archivedCaptures = append(archivedCaptures, archivedCapture)

	}
	archiveStatus := ArchiveStatus{
		Files:    archivedCaptures,
		NumFiles: numFiles,
		Size:     size,
	}

	archiveStats, marshalErr := json.Marshal(archiveStatus)
	if marshalErr != nil {
		s.logger.Error("Failed to marshal archive status", "error", marshalErr)
		jsonResponse(w, http.StatusInternalServerError, StatusRes{Status: "error"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(archiveStats)
}
