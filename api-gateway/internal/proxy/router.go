package proxy

import (
	"github.com/notes-in-the-cloud/notes-cloud-api-gateway/internal/middlewares"
	"github.com/notes-in-the-cloud/notes-cloud-api-gateway/internal/notifications"
	"github.com/notes-in-the-cloud/notes-cloud-api-gateway/internal/websocket"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/notes-in-the-cloud/notes-cloud-api-gateway/internal/probes"
	"github.com/notes-in-the-cloud/notes-cloud-jwt-utils/accesstoken"
)

type jwtValidator interface {
	ValidateAccessToken(rawToken string) (*accesstoken.Claims, error)
}

func NewRouter(
	p *proxy,
	jwtValidator jwtValidator,
	internalToken string,
	allowedOrigins []string,
) *mux.Router {
	r := mux.NewRouter()

	authMiddleware := accesstoken.AuthMiddleware(jwtValidator)
	wsAuthMiddleware := middlewares.AuthMiddlewareWithQueryToken(jwtValidator)

	wsHub := websocket.NewHub()
	wsHandler := websocket.NewHandler(wsHub, allowedOrigins)
	notificationsHandler := notifications.NewHandler(wsHub)

	// Health check endpoints
	r.HandleFunc("/api/healthz", probes.Healthz).Methods(http.MethodGet)
	r.HandleFunc("/api/readyz", probes.Readyz).Methods(http.MethodGet)

	// WebSocket endpoint.
	//
	// Browser WebSocket API cannot easily send Authorization headers,
	// so /ws supports ?token=<access_token>.
	wsRouter := r.NewRoute().Subrouter()
	wsRouter.Use(wsAuthMiddleware)
	wsRouter.HandleFunc("/ws", wsHandler.Connect).Methods(http.MethodGet)

	// Internal service-to-service endpoints.
	//
	// This must not be public. It is called by reminder-service.
	internal := r.PathPrefix("/internal").Subrouter()
	internal.Use(middlewares.InternalToken(internalToken))
	internal.HandleFunc(
		"/notifications/{userId}",
		notificationsHandler.Push,
	).Methods(http.MethodPost)

	api := r.PathPrefix("/api/v1").Subrouter()

	// ============ Auth endpoints (public) ============
	api.HandleFunc("/auth/register", p.Auth).Methods(http.MethodPost)
	api.HandleFunc("/auth/login", p.Auth).Methods(http.MethodPost)
	api.HandleFunc("/auth/logout", p.Auth).Methods(http.MethodPost)
	api.HandleFunc("/auth/refresh", p.Auth).Methods(http.MethodPost)
	api.HandleFunc("/auth/verify", p.Auth).Methods(http.MethodPost)
	api.HandleFunc("/auth/resend", p.Auth).Methods(http.MethodPost)
	api.HandleFunc("/auth/google/start", p.Auth).Methods(http.MethodGet)
	api.HandleFunc("/auth/google/callback", p.Auth).Methods(http.MethodGet)
	api.HandleFunc("/auth/gitlab/start", p.Auth).Methods(http.MethodGet)
	api.HandleFunc("/auth/gitlab/callback", p.Auth).Methods(http.MethodGet)

	// ============ Sharing (public) ============
	api.HandleFunc("/share-links/{token}", p.Sharing).Methods(http.MethodGet)

	// ============ Protected endpoints ============
	protected := api.PathPrefix("").Subrouter()
	protected.Use(authMiddleware)

	protected.HandleFunc("/me", p.Auth).Methods(http.MethodGet)

	protected.HandleFunc("/notes", p.Notes).Methods(http.MethodGet, http.MethodPost)
	protected.HandleFunc("/notes/{note_id}", p.Notes).Methods(http.MethodGet, http.MethodPut, http.MethodDelete)

	protected.HandleFunc("/notes/{note_id}/share-links", p.Sharing).Methods(http.MethodPost)

	protected.HandleFunc("/todos", p.Todo).Methods(http.MethodGet, http.MethodPost)
	protected.HandleFunc("/todos/{todo_id}", p.Todo).Methods(http.MethodGet, http.MethodPut, http.MethodDelete)

	protected.HandleFunc("/todo-lists", p.Todo).Methods(http.MethodGet, http.MethodPost)
	protected.HandleFunc("/todo-lists/{list_id}", p.Todo).Methods(http.MethodGet, http.MethodPut, http.MethodDelete)

	protected.HandleFunc("/reminders", p.Reminder).Methods(http.MethodGet, http.MethodPost, http.MethodPut)
	protected.HandleFunc("/reminders/{reminder_id}", p.Reminder).Methods(http.MethodGet, http.MethodDelete)

	protected.HandleFunc("/notifications", p.Reminder).Methods(http.MethodGet, http.MethodDelete)
	protected.HandleFunc("/notifications/unread-count", p.Reminder).Methods(http.MethodGet)
	protected.HandleFunc("/notifications/read-all", p.Reminder).Methods(http.MethodPost)
	protected.HandleFunc("/notifications/{notification_id}/read", p.Reminder).Methods(http.MethodPost)

	return r
}
