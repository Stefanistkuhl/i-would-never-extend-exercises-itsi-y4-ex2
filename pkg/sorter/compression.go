package sorter

import (
	"compress/gzip"
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/stefanistkuhl/i-would-never-extend-exercises-itsi-y4-ex2/pkg/db"
	"github.com/stefanistkuhl/i-would-never-extend-exercises-itsi-y4-ex2/pkg/db/sqlc"
)

func (s *Server) CompressFileHandler(w http.ResponseWriter, r *http.Request) {
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

	if capture.Compressed.Bool {
		jsonResponse(w, http.StatusBadRequest, StatusRes{Status: "error"})
		return
	}

	if capture.Archived.Bool {
		jsonResponse(w, http.StatusBadRequest, StatusRes{Status: "error"})
		return
	}

	err = s.compressFile(int(captureID), capture.FilePath, store)
	if err != nil {
		s.logger.Error("Failed to compress file", "error", err, "id", captureID)
		jsonResponse(w, http.StatusInternalServerError, StatusRes{Status: "error"})
		return
	}

	jsonResponse(w, http.StatusOK, StatusRes{Status: "ok"})
}

func (s *Server) CompressTriggerHandler(w http.ResponseWriter, r *http.Request) {
	store, err := db.InitIfNeeded()
	if err != nil {
		s.logger.Error("Failed to initialize database", "error", err)
		jsonResponse(w, http.StatusInternalServerError, StatusRes{Status: "error"})
		return
	}

	rows, err := store.GetPendingCompressions(context.Background(), 100)
	if err != nil {
		s.logger.Error("Failed to get pending compressions", "error", err)
		jsonResponse(w, http.StatusInternalServerError, StatusRes{Status: "error"})
		return
	}

	result := CompressionTriggerRes{
		Processed: 0,
		Failed:    0,
		Errors:    []string{},
	}

	for _, row := range rows {
		err := s.compressFile(int(row.ID), row.FilePath, store)
		if err != nil {
			result.Failed++
			errMsg := fmt.Sprintf("Failed to compress file %s (id: %d): %v", row.FilePath, row.ID, err)
			s.logger.Error(errMsg)
			result.Errors = append(result.Errors, errMsg)
			continue
		}
		result.Processed++
	}

	jsonResponse(w, http.StatusOK, result)
}

func (s *Server) compressFile(id int, filePath string, store *db.Store) error {
	cfg := s.GetConfig()

	if cfg.LogLevel == "info" {
		s.logger.Info("Compressing file", "path", filePath, "id", id)
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", filePath)
	}

	if strings.HasSuffix(filePath, ".gz") {
		if cfg.LogLevel == "info" {
			s.logger.Info("File already compressed", "path", filePath)
		}
		return nil
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
		s.logger.Error("Failed to remove original file after compression", "path", filePath, "error", removeErr)
	}

	updateErr := store.UpdateFilePath(context.Background(), sqlc.UpdateFilePathParams{
		FilePath: compressedPath,
		ID:       int64(id),
	})
	if updateErr != nil {
		s.logger.Error("Failed to update file path in database", "error", updateErr)
	}

	markErr := store.MarkCaptureAsCompressed(context.Background(), int64(id))
	if markErr != nil {
		return fmt.Errorf("failed to mark capture as compressed: %w", markErr)
	}

	return nil
}
