package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"

	"travel-collab/backend/internal/config"
	"travel-collab/backend/internal/ws"
)

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	tripID := chi.URLParam(r, "trip_id")
	token := r.URL.Query().Get("token")
	if token == "" {
		writeError(w, http.StatusUnauthorized, "missing token")
		return
	}
	claims, err := s.auth.ParseToken(token)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid token")
		return
	}
	ok, err := s.store.IsTripMember(r.Context(), tripID, claims.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to check access")
		return
	}
	if !ok {
		writeError(w, http.StatusForbidden, "you are not a member of this trip")
		return
	}

	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return isWebSocketOriginAllowed(s.cfg, r)
		},
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	client := &ws.Client{
		Hub:    s.hub,
		Store:  s.store,
		Conn:   conn,
		Send:   make(chan []byte, 256),
		TripID: tripID,
		UserID: claims.UserID,
	}
	s.hub.Register(client)
	go client.WritePump()
	go client.ReadPump()
}

func isWebSocketOriginAllowed(cfg config.Config, r *http.Request) bool {
	origin := r.Header.Get("Origin")
	return origin == "" || cfg.IsOriginAllowed(origin)
}
