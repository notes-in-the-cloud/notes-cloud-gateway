package middlewares

import (
	"net/http"
)

const InternalTokenHeader = "X-Internal-Token"

func InternalToken(requiredToken string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if requiredToken == "" {
				http.Error(w, "internal token is not configured", http.StatusInternalServerError)
				return
			}

			providedToken := r.Header.Get(InternalTokenHeader)
			if providedToken == "" {
				http.Error(w, "missing internal token", http.StatusUnauthorized)
				return
			}

			if providedToken != requiredToken {
				http.Error(w, "invalid internal token", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
