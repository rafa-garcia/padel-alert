package api

import "context"

// Context keys
type contextKey string

const (
	// contextKeyUserID is the key for the user ID in the request context
	contextKeyUserID contextKey = "user_id"
)

// WithUserID adds a user ID to the context
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, contextKeyUserID, userID)
}

// GetUserID gets a user ID from the context
func GetUserID(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(contextKeyUserID).(string)
	return userID, ok
}
