package sorter

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/stefanistkuhl/i-would-never-extend-exercises-itsi-y4-ex2/pkg/db"
)

func (s *Server) ExportStoreHandler(w http.ResponseWriter, r *http.Request) {
	cfg := s.GetConfig()

	tmpFile, err := os.CreateTemp("", "pcapstore-export-*.tar.gz")
	if err != nil {
		s.logger.Error("Failed to create temporary file", "error", err)
		jsonResponse(w, http.StatusInternalServerError, StatusRes{Status: "error"})
		return
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	gzWriter := gzip.NewWriter(tmpFile)
	defer gzWriter.Close()

	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	store, dbErr := db.InitIfNeeded()
	if dbErr != nil {
		s.logger.Error("Failed to initialize database", "error", dbErr)
		jsonResponse(w, http.StatusInternalServerError, StatusRes{Status: "error"})
		return
	}

	cwd, err := os.Getwd()
	if err != nil {
		s.logger.Error("Failed to get current directory", "error", err)
		jsonResponse(w, http.StatusInternalServerError, StatusRes{Status: "error"})
		return
	}
	dbPath := filepath.Join(cwd, "pcapStore.db")
	if _, err := os.Stat(dbPath); err == nil {
		if err := addFileToTar(tarWriter, dbPath, "pcapStore.db"); err != nil {
			s.logger.Error("Failed to add database to archive", "error", err)
			jsonResponse(w, http.StatusInternalServerError, StatusRes{Status: "error"})
			return
		}
	} else {
		s.logger.Warn("Database file not found", "path", dbPath)
	}

	captures, err := store.GetCaptures(context.Background())
	if err != nil {
		s.logger.Error("Failed to get captures", "error", err)
		jsonResponse(w, http.StatusInternalServerError, StatusRes{Status: "error"})
		return
	}

	addedFiles := make(map[string]bool)
	for _, capture := range captures {
		filePath := capture.FilePath
		if filePath == "" {
			continue
		}

		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			s.logger.Warn("File does not exist", "path", filePath)
			continue
		}

		if addedFiles[filePath] {
			continue
		}
		addedFiles[filePath] = true

		relPath, err := filepath.Rel(cfg.OrganizedDir, filePath)
		if err != nil {
			relPath, err = filepath.Rel(cfg.ArchiveDir, filePath)
			if err != nil {
				relPath = filepath.Base(filePath)
			} else {
				relPath = filepath.Join("archive", relPath)
			}
		} else {
			relPath = filepath.Join("organized", relPath)
		}

		if err := addFileToTar(tarWriter, filePath, relPath); err != nil {
			s.logger.Warn("Failed to add file to archive", "path", filePath, "error", err)
			continue
		}
	}

	if err := addDirectoryToTar(tarWriter, cfg.ArchiveDir, "archive"); err != nil {
		s.logger.Warn("Failed to add archive directory", "error", err)
	}

	tarWriter.Close()
	gzWriter.Close()
	tmpFile.Close()

	stat, err := os.Stat(tmpFile.Name())
	if err != nil {
		s.logger.Error("Failed to stat archive file", "error", err)
		jsonResponse(w, http.StatusInternalServerError, StatusRes{Status: "error"})
		return
	}

	archiveFile, err := os.Open(tmpFile.Name())
	if err != nil {
		s.logger.Error("Failed to open archive file", "error", err)
		jsonResponse(w, http.StatusInternalServerError, StatusRes{Status: "error"})
		return
	}
	defer archiveFile.Close()

	filename := fmt.Sprintf("pcapstore-export-%s.tar.gz", time.Now().Format("20060102-150405"))
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	w.Header().Set("Content-Type", "application/gzip")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", stat.Size()))

	_, err = io.Copy(w, archiveFile)
	if err != nil {
		s.logger.Error("Failed to stream archive", "error", err)
		return
	}
}

func addFileToTar(tarWriter *tar.Writer, filePath, tarPath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return err
	}

	header := &tar.Header{
		Name:    tarPath,
		Size:    stat.Size(),
		Mode:    int64(stat.Mode()),
		ModTime: stat.ModTime(),
	}

	if err := tarWriter.WriteHeader(header); err != nil {
		return err
	}

	_, err = io.Copy(tarWriter, file)
	return err
}

func addDirectoryToTar(tarWriter *tar.Writer, dirPath, tarPrefix string) error {
	return filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(dirPath, path)
		if err != nil {
			return err
		}

		tarPath := filepath.Join(tarPrefix, relPath)
		return addFileToTar(tarWriter, path, tarPath)
	})
}
