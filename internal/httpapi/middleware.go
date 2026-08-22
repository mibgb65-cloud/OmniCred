package httpapi

import (
	"log/slog"
	"net/http"
	"slices"
	"time"
)

var desktopOrigins = []string{"http://wails.localhost", "https://wails.localhost"}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (writer *statusWriter) WriteHeader(status int) {
	writer.status = status
	writer.ResponseWriter.WriteHeader(status)
}

func withRequestLogging(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		started := time.Now()
		writer := &statusWriter{ResponseWriter: response, status: http.StatusOK}
		next.ServeHTTP(writer, request)
		logger.Info("http request",
			"method", request.Method,
			"path", request.URL.Path,
			"status", writer.status,
			"duration", time.Since(started),
		)
	})
}

func withRecovery(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				logger.Error("http panic", "method", request.Method, "path", request.URL.Path)
				writeError(response, http.StatusInternalServerError, "internal_error", "an internal error occurred")
			}
		}()
		next.ServeHTTP(response, request)
	})
}

func withSecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		response.Header().Set("X-Content-Type-Options", "nosniff")
		response.Header().Set("Referrer-Policy", "no-referrer")
		if request.URL.Path == "/api/v1" || len(request.URL.Path) > 8 && request.URL.Path[:8] == "/api/v1/" {
			response.Header().Set("Cache-Control", "no-store")
		}
		next.ServeHTTP(response, request)
	})
}

func withDesktopCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		origin := request.Header.Get("Origin")
		if slices.Contains(desktopOrigins, origin) {
			response.Header().Set("Access-Control-Allow-Origin", origin)
			response.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			response.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			response.Header().Add("Vary", "Origin")
			if request.Method == http.MethodOptions {
				response.WriteHeader(http.StatusNoContent)
				return
			}
		}
		next.ServeHTTP(response, request)
	})
}
