package middleware

import (
	"net/http"
	"strings"
)

func CORS(frontendURL string, allowedOrigins []string) func(http.Handler) http.Handler {
	allowed := make(map[string]struct{})
	add := func(origin string) {
		cleaned := strings.TrimRight(strings.TrimSpace(origin), "/")
		if cleaned != "" {
			allowed[cleaned] = struct{}{}
		}
	}

	add(frontendURL)
	for _, origin := range allowedOrigins {
		add(origin)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := strings.TrimRight(r.Header.Get("Origin"), "/")
			if origin != "" {
				if _, ok := allowed["*"]; ok {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					w.Header().Set("Vary", "Origin")
				} else if _, ok := allowed[origin]; ok || isLocalDevelopmentOrigin(origin) {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					w.Header().Set("Vary", "Origin")
				}
			}
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func isLocalDevelopmentOrigin(origin string) bool {
	return strings.HasPrefix(origin, "capacitor://localhost") ||
		strings.HasPrefix(origin, "http://localhost") ||
		strings.HasPrefix(origin, "http://127.0.0.1") ||
		strings.HasPrefix(origin, "http://192.168.") ||
		strings.HasPrefix(origin, "http://10.") ||
		strings.HasPrefix(origin, "http://172.16.") ||
		strings.HasPrefix(origin, "http://172.17.") ||
		strings.HasPrefix(origin, "http://172.18.") ||
		strings.HasPrefix(origin, "http://172.19.") ||
		strings.HasPrefix(origin, "http://172.20.") ||
		strings.HasPrefix(origin, "http://172.21.") ||
		strings.HasPrefix(origin, "http://172.22.") ||
		strings.HasPrefix(origin, "http://172.23.") ||
		strings.HasPrefix(origin, "http://172.24.") ||
		strings.HasPrefix(origin, "http://172.25.") ||
		strings.HasPrefix(origin, "http://172.26.") ||
		strings.HasPrefix(origin, "http://172.27.") ||
		strings.HasPrefix(origin, "http://172.28.") ||
		strings.HasPrefix(origin, "http://172.29.") ||
		strings.HasPrefix(origin, "http://172.30.") ||
		strings.HasPrefix(origin, "http://172.31.")
}
