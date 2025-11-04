package sorter

import (
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	"github.com/stefanistkuhl/i-would-never-extend-exercises-itsi-y4-ex2/pkg/config"
	"github.com/stefanistkuhl/i-would-never-extend-exercises-itsi-y4-ex2/pkg/logger"
)

const socket = "/tmp/pcap-sorter.sock"

type Server struct {
	logger   logger.Logger
	cfg      config.Config
	password string
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
		r.Get("/files", s.GetFiles)
		r.Get("/file/{id}", s.GetFile)
		r.Delete("/file/{id}", s.DeleteFile)
	}

	r.Route("/api", func(r chi.Router) {
		statusRoutes(r)
		fileRoutes(r)
	})

	// Archiving Endpoints
	r.Get("/archive", s.HandleArchive)

	if cfg.LogLevel == "info" {
		s.logger.Info("HTTP Server listening", "address", listenAddr)
	}

	server := &http.Server{Handler: r}
	listenErr := server.Serve(listener)
	if listenErr != nil {
		s.logger.Fatal("Failed to start HTTP server", "error", listenErr)
	}
}
