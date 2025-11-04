package sorter

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	"github.com/stefanistkuhl/i-would-never-extend-exercises-itsi-y4-ex2/pkg/config"
	"github.com/stefanistkuhl/i-would-never-extend-exercises-itsi-y4-ex2/pkg/logger"
)

const socket = "/tmp/pcap-sorter.sock"

type Server struct {
	logger   logger.Logger
	cfg      config.Config
	cfgMu    sync.RWMutex
	password string
}

func (s *Server) GetConfig() config.Config {
	s.cfgMu.RLock()
	defer s.cfgMu.RUnlock()
	return s.cfg
}

func (s *Server) UpdateConfig(cfg config.Config) {
	s.cfgMu.Lock()
	defer s.cfgMu.Unlock()
	s.cfg = cfg
}

func startHTTPServer(lg logger.Logger, cfg config.Config) {
	s := &Server{
		logger: lg,
		cfg:    cfg,
	}

	r := chi.NewRouter()
	var listener net.Listener
	var listenAddr string
	var err error

	if cfg.ExposeService {
		envErr := godotenv.Load()
		if envErr != nil {
			s.logger.Warn("Failed to load .env file, falling back to environment variables")
		}

		s.password = os.Getenv("SORTER_PASSWORD")
		if s.password == "" {
			s.logger.Fatal("SORTER_PASSWORD environment variable is not set")
		}

		r.Use(s.simpleAuthMiddleware)
		listenAddr = fmt.Sprintf(":%d", cfg.Port)
		listener, err = net.Listen("tcp", listenAddr)
	} else {
		listenAddr = socket
		os.Remove(socket)
		listener, err = net.Listen("unix", socket)
	}

	if err != nil {
		s.logger.Fatal("Failed to create listener", "error", err)
	}

	// Health & Status Endpoints
	statusRoutes := func(r chi.Router) {
		r.Get("/health", s.HealthHandler)
		r.Get("/status", s.StatusHandler)
		r.Get("/version", s.VersionHandler)
	}

	// File Endpoints
	fileRoutes := func(r chi.Router) {
		r.Get("/files/{id}/download", s.FileDownloadHandler)
		r.Get("/files/{id}/stats", s.GetFileStatsHandler)
		r.Get("/files", s.GetFilesHandler)
		r.Get("/file/{id}", s.GetFileHandler)
		r.Delete("/file/{id}", s.DeleteFileHandler)
	}

	archiveRoutes := func(r chi.Router) {
		r.Get("/archive", s.GetArchiveHandler)
		r.Post("/archive/{id}", s.ArchiveFileHandler)
		r.Get("/archive/status", s.ArchiveStatusHandler)
	}

	// Config Endpoints
	configRoutes := func(r chi.Router) {
		r.Get("/config", s.GetConfigHandler)
		r.Put("/config", s.UpdateConfigHandler)
	}

	// Cleanup Endpoints
	cleanupRoutes := func(r chi.Router) {
		r.Get("/cleanup/candidates", s.GetCleanupCandidatesHandler)
		r.Post("/cleanup/execute", s.CleanupExecuteHandler)
	}

	// Statistics Endpoints
	statsRoutes := func(r chi.Router) {
		r.Get("/summary", s.GetSummaryHandler)
		r.Get("/stats/by-hostname", s.GetStatsByHostnameHandler)
		r.Get("/stats/by-scenario", s.GetStatsByScenarioHandler)
	}

	// Compression Endpoints
	compressionRoutes := func(r chi.Router) {
		r.Post("/compression/{id}", s.CompressFileHandler)
		r.Post("/compression/trigger", s.CompressTriggerHandler)
	}

	// Search & Query Endpoints
	searchRoutes := func(r chi.Router) {
		r.Get("/search", s.SearchHandler)
		r.Get("/files/by-hostname/{host}", s.GetFilesByHostnameHandler)
		r.Get("/files/by-scenario/{scenario}", s.GetFilesByScenarioHandler)
		r.Post("/query", s.QuerySQLHandler)
	}

	r.Route("/api", func(r chi.Router) {
		statusRoutes(r)
		fileRoutes(r)
		archiveRoutes(r)
		configRoutes(r)
		cleanupRoutes(r)
		statsRoutes(r)
		compressionRoutes(r)
		searchRoutes(r)
	})

	if cfg.LogLevel == "info" {
		s.logger.Info("HTTP Server listening", "address", listenAddr)
	}

	server := &http.Server{Handler: r}
	listenErr := server.Serve(listener)
	if listenErr != nil {
		s.logger.Fatal("Failed to start HTTP server", "error", listenErr)
	}
}
