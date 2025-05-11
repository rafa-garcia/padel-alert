package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/rafa-garcia/padel-alert/internal/logger"
)

// HealthResponse represents the response from the health check endpoint
type HealthResponse struct {
	Status    string `json:"status"`
	Version   string `json:"version"`
	Timestamp string `json:"timestamp"`
}

// HealthHandler handles the health check endpoint
type HealthHandler struct {
	Version string
}

// HealthCheck returns the health status of the service
func (h *HealthHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	healthResp := HealthResponse{
		Status:    "OK",
		Version:   h.Version,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	resp := Response{
		Data:   healthResp,
		Error:  nil,
		Status: http.StatusOK,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		logger.Error("Failed to encode JSON response", err)
	}
}
