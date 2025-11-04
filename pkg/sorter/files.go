package sorter

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/stefanistkuhl/i-would-never-extend-exercises-itsi-y4-ex2/pkg/db"
)

func (s *Server) GetFiles(w http.ResponseWriter, r *http.Request) {
	store, dbErr := db.InitIfNeeded()
	if dbErr != nil {
		s.logger.Error("Failed to initialize database", "error", dbErr)
		jsonResponse(w, http.StatusInternalServerError, StatusRes{Status: "error"})
		return
	}
	captures, getCapturesErr := store.GetCaptures(context.Background())
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

func (s *Server) GetFile(w http.ResponseWriter, r *http.Request) {
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
	captureJSON, marshalErr := json.Marshal(capture)
	if marshalErr != nil {
		s.logger.Error("Failed to marshal capture", "error", marshalErr)
		jsonResponse(w, http.StatusInternalServerError, StatusRes{Status: "error"})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(captureJSON)
}

func (s *Server) DeleteFile(w http.ResponseWriter, r *http.Request) {
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

	deleteErr := store.DeleteCapture(context.Background(), captureID)
	if deleteErr != nil {
		s.logger.Error("Failed to delete capture", "error", deleteErr, "id", captureID)
		jsonResponse(w, http.StatusInternalServerError, StatusRes{Status: "error"})
		return
	}

	removeErr := os.Remove(capture.FilePath)
	if removeErr != nil {
		s.logger.Error("Failed to remove file", "error", removeErr, "path", capture.FilePath)
		jsonResponse(w, http.StatusInternalServerError, StatusRes{Status: "error"})
		return
	}

	jsonResponse(w, http.StatusOK, StatusRes{Status: "ok"})
}

func (s *Server) FileDownloadHandler(w http.ResponseWriter, r *http.Request) {
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

	filePath := capture.FilePath
	fileName := filepath.Base(filePath)

	file, err := os.Open(filePath)
	if err != nil {
		s.logger.Error("Failed to open file", "error", err, "path", filePath)
		if os.IsNotExist(err) {
			jsonResponse(w, http.StatusNotFound, StatusRes{Status: "error"})
		} else {
			jsonResponse(w, http.StatusInternalServerError, StatusRes{Status: "error"})
		}
		return
	}
	defer file.Close()

	fileStat, err := file.Stat()
	if err != nil {
		s.logger.Error("Failed to stat file", "error", err)
		jsonResponse(w, http.StatusInternalServerError, StatusRes{Status: "error"})
		return
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", strconv.FormatInt(fileStat.Size(), 10))

	http.ServeContent(w, r, fileName, fileStat.ModTime(), file)
}
