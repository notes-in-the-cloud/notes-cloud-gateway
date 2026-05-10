package websocket

import (
	"log"
	"net/http"
	"strings"

	gwebsocket "github.com/gorilla/websocket"
	"github.com/notes-in-the-cloud/notes-cloud-jwt-utils/accesstoken"
)

type Handler struct {
	hub            *Hub
	allowedOrigins map[string]struct{}
}

func NewHandler(hub *Hub, allowedOrigins []string) *Handler {
	originSet := make(map[string]struct{}, len(allowedOrigins))

	for _, origin := range allowedOrigins {
		normalizedOrigin := strings.ToLower(strings.TrimSpace(origin))
		if normalizedOrigin == "" {
			continue
		}

		originSet[normalizedOrigin] = struct{}{}
	}

	return &Handler{
		hub:            hub,
		allowedOrigins: originSet,
	}
}

func (h *Handler) Connect(w http.ResponseWriter, r *http.Request) {
	userID, err := accesstoken.UserIDFromContext(r.Context())
	if err != nil {
		accesstoken.WriteErrorResponse(
			w,
			http.StatusUnauthorized,
			accesstoken.ErrInvalidToken,
			"user is not authenticated",
		)
		return
	}

	upgrader := gwebsocket.Upgrader{
		CheckOrigin:     h.checkOrigin,
		ReadBufferSize:  1024,
		WriteBufferSize: 4096,
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("ws: upgrade failed for user=%s: %v", userID, err)
		return
	}

	h.hub.Register(userID, conn)
}

func (h *Handler) checkOrigin(r *http.Request) bool {
	if len(h.allowedOrigins) == 0 {
		return true
	}

	origin := strings.ToLower(strings.TrimSpace(r.Header.Get("Origin")))

	_, allowed := h.allowedOrigins[origin]

	return allowed
}
