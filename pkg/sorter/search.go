package sorter

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stefanistkuhl/i-would-never-extend-exercises-itsi-y4-ex2/pkg/db"
	"github.com/stefanistkuhl/i-would-never-extend-exercises-itsi-y4-ex2/pkg/db/sqlc"
)

func (s *Server) SearchHandler(w http.ResponseWriter, r *http.Request) {
	store, err := db.InitIfNeeded()
	if err != nil {
		s.logger.Error("Failed to initialize database", "error", err)
		jsonResponse(w, http.StatusInternalServerError, StatusRes{Status: "error"})
		return
	}

	hostname := r.URL.Query().Get("hostname")
	scenario := r.URL.Query().Get("scenario")
	archivedParam := r.URL.Query().Get("archived")
	compressedParam := r.URL.Query().Get("compressed")

	allCaptures, err := store.GetCaptures(context.Background())
	if err != nil {
		s.logger.Error("Failed to get captures", "error", err)
		jsonResponse(w, http.StatusInternalServerError, StatusRes{Status: "error"})
		return
	}

	var filtered []sqlc.Capture
	for _, capture := range allCaptures {
		if hostname != "" && capture.Hostname != hostname {
			continue
		}
		if scenario != "" && capture.Scenario != scenario {
			continue
		}
		if archivedParam != "" {
			archived, err := strconv.ParseBool(archivedParam)
			if err == nil {
				if capture.Archived.Bool != archived {
					continue
				}
			}
		}
		if compressedParam != "" {
			compressed, err := strconv.ParseBool(compressedParam)
			if err == nil {
				if capture.Compressed.Bool != compressed {
					continue
				}
			}
		}
		filtered = append(filtered, capture)
	}

	results := make([]SearchResult, 0, len(filtered))
	for _, capture := range filtered {
		result := SearchResult{
			ID:              capture.ID,
			Hostname:        capture.Hostname,
			Scenario:        capture.Scenario,
			CaptureDatetime: capture.CaptureDatetime.Format(time.RFC3339),
			FilePath:        capture.FilePath,
			FileSize:        capture.FileSize,
			Compressed:      capture.Compressed.Bool,
			Archived:        capture.Archived.Bool,
		}
		if capture.CreatedAt.Valid {
			result.CreatedAt = capture.CreatedAt.Time.Format(time.RFC3339)
		}
		if capture.UpdatedAt.Valid {
			result.UpdatedAt = capture.UpdatedAt.Time.Format(time.RFC3339)
		}
		results = append(results, result)
	}

	jsonResponse(w, http.StatusOK, SearchRes{
		Results: results,
		Count:   len(results),
	})
}

func (s *Server) GetFilesByHostnameHandler(w http.ResponseWriter, r *http.Request) {
	hostname := chi.URLParam(r, "host")
	if hostname == "" {
		jsonResponse(w, http.StatusBadRequest, StatusRes{Status: "error"})
		return
	}

	store, err := db.InitIfNeeded()
	if err != nil {
		s.logger.Error("Failed to initialize database", "error", err)
		jsonResponse(w, http.StatusInternalServerError, StatusRes{Status: "error"})
		return
	}

	captures, err := store.GetCapturesByHostname(context.Background(), hostname)
	if err != nil {
		s.logger.Error("Failed to get captures by hostname", "error", err, "hostname", hostname)
		jsonResponse(w, http.StatusInternalServerError, StatusRes{Status: "error"})
		return
	}

	results := make([]SearchResult, 0, len(captures))
	for _, capture := range captures {
		result := SearchResult{
			ID:              capture.ID,
			Hostname:        capture.Hostname,
			Scenario:        capture.Scenario,
			CaptureDatetime: capture.CaptureDatetime.Format(time.RFC3339),
			FilePath:        capture.FilePath,
			FileSize:        capture.FileSize,
			Compressed:      capture.Compressed.Bool,
			Archived:        capture.Archived.Bool,
		}
		if capture.CreatedAt.Valid {
			result.CreatedAt = capture.CreatedAt.Time.Format(time.RFC3339)
		}
		if capture.UpdatedAt.Valid {
			result.UpdatedAt = capture.UpdatedAt.Time.Format(time.RFC3339)
		}
		results = append(results, result)
	}

	jsonResponse(w, http.StatusOK, SearchRes{
		Results: results,
		Count:   len(results),
	})
}

func (s *Server) GetFilesByScenarioHandler(w http.ResponseWriter, r *http.Request) {
	scenario := chi.URLParam(r, "scenario")
	if scenario == "" {
		jsonResponse(w, http.StatusBadRequest, StatusRes{Status: "error"})
		return
	}

	store, err := db.InitIfNeeded()
	if err != nil {
		s.logger.Error("Failed to initialize database", "error", err)
		jsonResponse(w, http.StatusInternalServerError, StatusRes{Status: "error"})
		return
	}

	captures, err := store.GetCapturesByScenario(context.Background(), scenario)
	if err != nil {
		s.logger.Error("Failed to get captures by scenario", "error", err, "scenario", scenario)
		jsonResponse(w, http.StatusInternalServerError, StatusRes{Status: "error"})
		return
	}

	results := make([]SearchResult, 0, len(captures))
	for _, capture := range captures {
		result := SearchResult{
			ID:              capture.ID,
			Hostname:        capture.Hostname,
			Scenario:        capture.Scenario,
			CaptureDatetime: capture.CaptureDatetime.Format(time.RFC3339),
			FilePath:        capture.FilePath,
			FileSize:        capture.FileSize,
			Compressed:      capture.Compressed.Bool,
			Archived:        capture.Archived.Bool,
		}
		if capture.CreatedAt.Valid {
			result.CreatedAt = capture.CreatedAt.Time.Format(time.RFC3339)
		}
		if capture.UpdatedAt.Valid {
			result.UpdatedAt = capture.UpdatedAt.Time.Format(time.RFC3339)
		}
		results = append(results, result)
	}

	jsonResponse(w, http.StatusOK, SearchRes{
		Results: results,
		Count:   len(results),
	})
}

func (s *Server) QuerySQLHandler(w http.ResponseWriter, r *http.Request) {
	var req SQLQueryReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.logger.Error("Failed to decode request", "error", err)
		jsonResponse(w, http.StatusBadRequest, StatusRes{Status: "error"})
		return
	}
	defer r.Body.Close()

	if req.Query == "" {
		jsonResponse(w, http.StatusBadRequest, StatusRes{Status: "error"})
		return
	}

	store, err := db.InitIfNeeded()
	if err != nil {
		s.logger.Error("Failed to initialize database", "error", err)
		jsonResponse(w, http.StatusInternalServerError, StatusRes{Status: "error"})
		return
	}

	s.logger.Info("Executing SQL query", "query", req.Query)

	rows, err := store.QueryContext(context.Background(), req.Query)
	if err != nil {
		s.logger.Error("Failed to execute SQL query", "error", err)
		jsonResponse(w, http.StatusBadRequest, SQLQueryRes{
			Success: false,
			Error:   err.Error(),
		})
		return
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		s.logger.Error("Failed to get columns", "error", err)
		jsonResponse(w, http.StatusInternalServerError, StatusRes{Status: "error"})
		return
	}

	var results []map[string]any
	for rows.Next() {
		values := make([]any, len(columns))
		valuePtrs := make([]any, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			s.logger.Error("Failed to scan row", "error", err)
			continue
		}

		rowMap := make(map[string]any)
		for i, col := range columns {
			val := values[i]
			if val == nil {
				rowMap[col] = nil
			} else {
				switch v := val.(type) {
				case []byte:
					rowMap[col] = string(v)
				case time.Time:
					rowMap[col] = v.Format(time.RFC3339)
				default:
					rowMap[col] = v
				}
			}
		}
		results = append(results, rowMap)
	}

	if err := rows.Err(); err != nil {
		s.logger.Error("Error iterating rows", "error", err)
		jsonResponse(w, http.StatusInternalServerError, StatusRes{Status: "error"})
		return
	}

	jsonResponse(w, http.StatusOK, SQLQueryRes{
		Success: true,
		Columns: columns,
		Rows:    results,
		Count:   len(results),
	})
}
