package middlewares

import (
	"github.com/gorilla/handlers"
	"net/http"
)

func CORS(next http.Handler, allowedOrigins []string) http.Handler {
	return handlers.CORS(
		handlers.AllowedOrigins(allowedOrigins),
		handlers.AllowedMethods([]string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodDelete,
			http.MethodOptions,
		}),
		handlers.AllowedHeaders([]string{
			"Accept",
			"Authorization",
			"Content-Type",
			InternalTokenHeader,
		}),
		handlers.AllowCredentials(), // Allow cookies to be sent
		handlers.OptionStatusCode(http.StatusNoContent),
	)(next)
}
