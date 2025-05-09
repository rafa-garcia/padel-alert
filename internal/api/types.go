package api

// Response represents a standard API response
type Response struct {
	Data    interface{} `json:"data,omitempty"`
	Error   *string     `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
	Status  int         `json:"status"`
}
