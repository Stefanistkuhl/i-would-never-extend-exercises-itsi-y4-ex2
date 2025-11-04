package sorter

import "net/http"

func (s *Server) HandleArchive(w http.ResponseWriter, r *http.Request) {
	if s.cfg.LogLevel == "info" {
		s.logger.Info("Archive endpoint accessed")
	}
	w.Write([]byte("archive"))
}
