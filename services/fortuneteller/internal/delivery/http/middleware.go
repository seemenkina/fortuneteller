package http

import (
	"context"
	"net/http"
)

// userToken is a context key for user token inside request context.
// We use private type to avoid collisions in context keys.
type userToken struct{}

// tokenFromReq extracts token (which was extracted from the cookie) from the request.
func tokenFromReq(r *http.Request) string {

	return r.Context().Value(userToken{}).(string)
}

// AuthenticatedUser ensures user is logged in.
func AuthenticatedUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("tokencookie")
		if err != nil || cookie.Value == "" {
			writeError(w, http.StatusUnauthorized, "cookie is empty: %v", err)
			return
		}

		next.ServeHTTP(w,
			r.WithContext(context.WithValue(r.Context(), userToken{}, cookie.Value)),
		)
	})
}
