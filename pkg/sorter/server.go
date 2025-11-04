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

var password string

func startHTTPServer(lg logger.Logger, cfg config.Config) {
	r := chi.NewRouter()
	var listener net.Listener
	var listenAddr string
	var err error

	if cfg.ExposeService {
		envErr := godotenv.Load()
		if envErr != nil {
			lg.Warn("Failed to load .env file, falling back to environment variables")
		}

		password = os.Getenv("SORTER_PASSWORD")
		if password == "" {
			lg.Fatal("SORTER_PASSWORD environment variable is not set")
		}

		r.Use(simpleAuthMiddleware)
		listenAddr = fmt.Sprintf(":%d", cfg.Port)
		listener, err = net.Listen("tcp", listenAddr)
	} else {
		listenAddr = socket
		os.Remove(socket)
		listener, err = net.Listen("unix", socket)
	}

	if err != nil {
		lg.Fatal("Failed to create listener", "error", err)
	}

	r.Get("/archive", handleArchive)

	if cfg.LogLevel == "info" {
		lg.Info("HTTP Server listening", "address", listenAddr)
	}

	server := &http.Server{Handler: r}
	listenErr := server.Serve(listener)
	if listenErr != nil {
		lg.Fatal("Failed to start HTTP server", "error", listenErr)
	}
}

func handleArchive(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("archive"))
}

func simpleAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token != password {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(http.StatusText(http.StatusUnauthorized)))
			return
		}
		next.ServeHTTP(w, r)
	})
}
