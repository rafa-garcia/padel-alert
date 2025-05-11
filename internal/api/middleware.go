package api

import (
	"context"
	"encoding/json"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/go-chi/cors"
	"github.com/rafa-garcia/padel-alert/internal/logger"
	"github.com/rafa-garcia/padel-alert/internal/metrics"
)

// ctxKey is a custom type for context keys to avoid collisions
type ctxKey int

const (
	RequestIDKey ctxKey = iota
)

// APIKeyAuth is a middleware that checks for a valid API key
func APIKeyAuth(validAPIKeys []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check header
			key := r.Header.Get("X-API-Key")

			// If not in header, check query param
			if key == "" {
				key = r.URL.Query().Get("api_key")
			}

			// Validate key
			if key == "" || !isValidAPIKey(key, validAPIKeys) {
				logger.Warn("Unauthorized API request", "ip", r.RemoteAddr, "path", r.URL.Path)
				metrics.HttpRequestsTotal.WithLabelValues(r.Method, r.URL.Path, "401").Inc()
				respondWithError(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Key is valid, proceed
			next.ServeHTTP(w, r)
		})
	}
}

// RequestLogger logs information about incoming requests and their duration
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a wrapped response writer that captures the status code
		ww := NewWrapResponseWriter(w)

		// Process the request
		next.ServeHTTP(ww, r)

		// Log the request
		duration := time.Since(start)
		statusCode := ww.Status()

		// Update metrics
		metrics.HttpRequestsTotal.WithLabelValues(r.Method, r.URL.Path, http.StatusText(statusCode)).Inc()
		metrics.HttpRequestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(duration.Seconds())

		// Log request details
		logger.Info("HTTP request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", statusCode,
			"duration", duration.String(),
			"size", ww.BytesWritten(),
			"ip", r.RemoteAddr,
		)
	})
}

// RequestID adds a unique request ID to the context
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get request ID from header or generate a new one
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = newRequestID()
		}

		// Add to response header
		w.Header().Set("X-Request-ID", requestID)

		// Add to context
		ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Recoverer is a middleware that recovers from panics
func Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rvr := recover(); rvr != nil {
				stackTrace := debug.Stack()

				// Log the error
				logger.Error("Panic recovered in HTTP handler",
					nil,
					"error", rvr,
					"stack", string(stackTrace),
					"path", r.URL.Path,
					"method", r.Method,
				)

				// Respond with error
				respondWithError(w, "Internal server error", http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// CORS returns a middleware that handles CORS
func CORS() func(http.Handler) http.Handler {
	return cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"}, // Adjust this in production
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-API-Key"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not freely browsable in seconds
	})
}

// SecurityHeaders adds security-related HTTP headers
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add common security headers
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Content-Security-Policy", "default-src 'self'")

		next.ServeHTTP(w, r)
	})
}

// isValidAPIKey checks if the provided key is in the list of valid keys
func isValidAPIKey(key string, validAPIKeys []string) bool {
	for _, validKey := range validAPIKeys {
		if key == validKey {
			return true
		}
	}
	return false
}

// respondWithError sends an error response
func respondWithError(w http.ResponseWriter, message string, status int) {
	msg := message
	resp := Response{
		Data:   nil,
		Error:  &msg,
		Status: status,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		logger.Error("Failed to encode JSON response", err)
	}
}

// respondWithJSON sends a JSON response
func respondWithJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(Response{
		Status: http.StatusOK,
		Data:   data,
	}); err != nil {
		logger.Error("Failed to encode JSON response", err)
	}
}

// respondWithSuccess sends a success message response
func respondWithSuccess(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(Response{
		Status:  http.StatusOK,
		Message: message,
	}); err != nil {
		logger.Error("Failed to encode JSON response", err)
	}
}

// newRequestID generates a unique request ID
func newRequestID() string {
	// TODO: use a more robust solution like UUID
	return time.Now().Format("20060102150405.000000")
}

// WrapResponseWriter is a wrapper around http.ResponseWriter that captures the status code
type WrapResponseWriter struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int
}

// NewWrapResponseWriter creates a new WrapResponseWriter
func NewWrapResponseWriter(w http.ResponseWriter) *WrapResponseWriter {
	return &WrapResponseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK, // Default status code
	}
}

// Status returns the HTTP status code
func (w *WrapResponseWriter) Status() int {
	return w.statusCode
}

// BytesWritten returns the number of bytes written
func (w *WrapResponseWriter) BytesWritten() int {
	return w.bytesWritten
}

// WriteHeader captures the status code
func (w *WrapResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

// Write captures the number of bytes written
func (w *WrapResponseWriter) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	w.bytesWritten += n
	return n, err
}
