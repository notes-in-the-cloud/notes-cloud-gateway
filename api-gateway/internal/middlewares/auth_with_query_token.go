package middlewares

import (
	"github.com/gorilla/mux"
	"github.com/notes-in-the-cloud/notes-cloud-jwt-utils/accesstoken"
	"net/http"
)

type jwtValidator interface {
	ValidateAccessToken(rawToken string) (*accesstoken.Claims, error)
}

func AuthMiddlewareWithQueryToken(jwtValidator jwtValidator) mux.MiddlewareFunc {
	authMiddleware := accesstoken.AuthMiddleware(jwtValidator)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Authorization") == "" {
				token := r.URL.Query().Get("token")
				if token != "" {
					clonedRequest := r.Clone(r.Context())
					clonedRequest.Header = r.Header.Clone()
					clonedRequest.Header.Set("Authorization", "Bearer "+token)

					r = clonedRequest
				}
			}

			authMiddleware(next).ServeHTTP(w, r)
		})
	}
}
