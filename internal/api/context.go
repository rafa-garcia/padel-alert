package api

import "context"

// Context keys
type contextKey string

const (
	contextKeyUserID contextKey = "user_id"
)

// WithUserID adds a user ID to the context
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, contextKeyUserID, userID)
}

// GetUserID gets a user ID from the context, with safe type assertion
func GetUserID(ctx context.Context) (string, bool) {
	val := ctx.Value(contextKeyUserID)
	if val == nil {
		return "", false
	}

	userID, ok := val.(string)
	return userID, ok
}
