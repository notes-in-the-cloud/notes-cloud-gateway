package notifications

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	ws "github.com/notes-in-the-cloud/notes-cloud-api-gateway/internal/websocket"
)

const maxBodyBytes = 64 * 1024

type Payload struct {
	NotificationID string `json:"notificationId"`
	ReminderID     string `json:"reminderId"`
	Heading        string `json:"heading"`
	Message        string `json:"message"`
	Priority       string `json:"priority"`
	FiredAt        string `json:"firedAt"`
	Type           string `json:"type"`
}

type Handler struct {
	hub *ws.Hub
}

func NewHandler(hub *ws.Hub) *Handler {
	return &Handler{
		hub: hub,
	}
}

func (h *Handler) Push(w http.ResponseWriter, r *http.Request) {
	userID := mux.Vars(r)["userId"]
	if userID == "" {
		http.Error(w, "missing userId", http.StatusBadRequest)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	defer r.Body.Close()

	var payload Payload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("notifications: failed to marshal payload for user=%s: %v", userID, err)
		http.Error(w, "failed to encode notification", http.StatusInternalServerError)
		return
	}

	delivered := h.hub.Send(userID, data)
	if !delivered {
		log.Printf("notifications: user=%s offline or buffer full, skipped websocket push", userID)
	}

	w.WriteHeader(http.StatusNoContent)
}
