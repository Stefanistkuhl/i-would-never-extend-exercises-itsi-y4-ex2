package sorter

import "net/http"

func (s *Server) simpleAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token != s.password {
			s.logger.Warn("Unauthorized access attempt", "path", r.URL.Path, "ip", r.RemoteAddr)
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(http.StatusText(http.StatusUnauthorized)))
			return
		}
		next.ServeHTTP(w, r)
	})
}

