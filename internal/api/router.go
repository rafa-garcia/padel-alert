package api

import (
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rafa-garcia/padel-alert/internal/storage"
)

// NewRouter creates a new Chi router with the configured routes
func NewRouter(version string, apiKeys []string, ruleStorage storage.RuleStorage, userStorage storage.UserStorage) *chi.Mux {
	r := chi.NewRouter()

	// Common middleware - order matters
	r.Use(RequestID)                            // Add request ID
	r.Use(RequestLogger)                        // Log requests
	r.Use(Recoverer)                            // Recover from panics
	r.Use(middleware.Timeout(60 * time.Second)) // Timeout for requests
	r.Use(CORS())                               // Handle CORS
	r.Use(SecurityHeaders)                      // Add security headers
	r.Use(middleware.Heartbeat("/ping"))        // Simple ping endpoint

	// Create handlers
	healthHandler := &HealthHandler{
		Version: version,
	}

	searchHandler := NewSearchHandler()

	ruleHandler := NewRuleHandler(ruleStorage, userStorage)

	// Public routes
	r.Group(func(r chi.Router) {
		r.Get("/api/v1/health", healthHandler.HealthCheck)
		r.Get("/api/v1/search", searchHandler.Search)
	})

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(APIKeyAuth(apiKeys))

		// Protected metrics endpoint
		r.Handle("/metrics", promhttp.Handler()) // Prometheus metrics endpoint

		// Rules API endpoints
		r.Route("/api/v1/rules", func(r chi.Router) {
			r.Get("/", ruleHandler.ListRules)
			r.Post("/", ruleHandler.CreateRule)
			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", ruleHandler.GetRule)
				r.Put("/", ruleHandler.UpdateRule)
				r.Delete("/", ruleHandler.DeleteRule)
			})
		})
	})

	return r
}
