package api

// Response represents a standard API response
type Response struct {
	Data   interface{} `json:"data"`
	Error  *string     `json:"error"`
	Status int         `json:"status"`
}
