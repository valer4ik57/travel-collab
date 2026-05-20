package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	appmiddleware "travel-collab/backend/internal/middleware"
)

func (s *Server) handleListMessages(w http.ResponseWriter, r *http.Request) {
	tripID := chi.URLParam(r, "trip_id")
	userID, _ := appmiddleware.UserIDFromContext(r.Context())
	if !s.ensureTripMember(w, r, tripID, userID) {
		return
	}
	messages, err := s.store.ListMessages(r.Context(), tripID, 100)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list messages")
		return
	}
	writeJSON(w, http.StatusOK, messages)
}
