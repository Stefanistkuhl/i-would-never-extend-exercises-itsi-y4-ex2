package sorter

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/stefanistkuhl/i-would-never-extend-exercises-itsi-y4-ex2/pkg/config"
	"github.com/stefanistkuhl/i-would-never-extend-exercises-itsi-y4-ex2/pkg/db"
)

func (s *Server) GetConfigHandler(w http.ResponseWriter, r *http.Request) {
	store, dbErr := db.InitIfNeeded()
	if dbErr != nil {
		s.logger.Error("Failed to initialize database", "error", dbErr)
		jsonResponse(w, http.StatusInternalServerError, StatusRes{Status: "error"})
		return
	}
	config, getConfigErr := store.GetConfig(context.Background())
	if getConfigErr != nil {
		s.logger.Error("Failed to get config", "error", getConfigErr)
		jsonResponse(w, http.StatusInternalServerError, StatusRes{Status: "error"})
		return
	}
	configJSON, marshalErr := json.Marshal(config)
	if marshalErr != nil {
		s.logger.Error("Failed to marshal config", "error", marshalErr)
		jsonResponse(w, http.StatusInternalServerError, StatusRes{Status: "error"})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(configJSON)
}

func (s *Server) UpdateConfigHandler(w http.ResponseWriter, r *http.Request) {
	var cfg config.Config
	bodyRaw, readAllErr := io.ReadAll(r.Body)
	if readAllErr != nil {
		s.logger.Error("Failed to read request body", "error", readAllErr)
		jsonResponse(w, http.StatusBadRequest, StatusRes{Status: "error"})
		return
	}
	defer r.Body.Close()
	err := json.Unmarshal(bodyRaw, &cfg)
	if err != nil {
		s.logger.Error("Failed to decode config", "error", err)
		jsonResponse(w, http.StatusBadRequest, StatusRes{Status: "error"})
		return
	}

	store, dbErr := db.InitIfNeeded()
	if dbErr != nil {
		s.logger.Error("Failed to initialize database", "error", dbErr)
		jsonResponse(w, http.StatusInternalServerError, StatusRes{Status: "error"})
		return
	}

	updateErr := store.UpdateConfig(context.Background(), cfg.ToUpdateParams())
	if updateErr != nil {
		s.logger.Error("Failed to update config", "error", updateErr)
		jsonResponse(w, http.StatusInternalServerError, StatusRes{Status: "error"})
		return
	}
	writeErr := config.WriteConfig(cfg, "config.toml")
	if writeErr != nil {
		s.logger.Error("Failed to write config", "error", writeErr)
		jsonResponse(w, http.StatusInternalServerError, StatusRes{Status: "error"})
		return
	}

	oldCfg := s.GetConfig()
	portChanged := cfg.Port != oldCfg.Port
	exposeServiceChanged := cfg.ExposeService != oldCfg.ExposeService
	s.UpdateConfig(cfg)

	if portChanged || exposeServiceChanged {
		s.logger.Warn("Config updated but server restart required for changes to take effect",
			"port_changed", portChanged, "expose_service_changed", exposeServiceChanged)
	} else {
		s.logger.Info("Config updated and applied")
	}

	jsonResponse(w, http.StatusOK, StatusRes{Status: "ok"})
}
