package api

import (
	"net/http"
)

// UserIDMiddleware extracts user ID from query parameter and sets it in context
func UserIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.URL.Query().Get("user_id")

		if userID != "" {
			ctx := WithUserID(r.Context(), userID)
			r = r.WithContext(ctx)
		}

		next.ServeHTTP(w, r)
	})
}
