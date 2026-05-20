package api

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	appmiddleware "travel-collab/backend/internal/middleware"
	"travel-collab/backend/internal/models"
)

type locationRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Lat         float64 `json:"lat"`
	Lng         float64 `json:"lng"`
	Category    string  `json:"category"`
}

func (s *Server) handleListLocations(w http.ResponseWriter, r *http.Request) {
	tripID := chi.URLParam(r, "trip_id")
	userID, _ := appmiddleware.UserIDFromContext(r.Context())
	if !s.ensureTripMember(w, r, tripID, userID) {
		return
	}
	locations, err := s.store.ListLocations(r.Context(), tripID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list locations")
		return
	}
	writeJSON(w, http.StatusOK, locations)
}

func (s *Server) handleCreateLocation(w http.ResponseWriter, r *http.Request) {
	tripID := chi.URLParam(r, "trip_id")
	userID, _ := appmiddleware.UserIDFromContext(r.Context())
	if _, ok := s.ensureTripRole(w, r, tripID, userID, "owner", "editor"); !ok {
		return
	}
	var req locationRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	if !validateLocationRequest(w, req) {
		return
	}
	location, err := s.store.CreateLocation(r.Context(), tripID, userID, strings.TrimSpace(req.Name), req.Description, req.Lat, req.Lng, normalizeCategory(req.Category))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create location")
		return
	}
	s.hub.Broadcast(tripID, models.WSEvent{Type: "LOCATION_ADDED", Payload: location})
	writeJSON(w, http.StatusCreated, location)
}

func (s *Server) handleUpdateLocation(w http.ResponseWriter, r *http.Request) {
	tripID := chi.URLParam(r, "trip_id")
	locationID := chi.URLParam(r, "location_id")
	userID, _ := appmiddleware.UserIDFromContext(r.Context())
	if _, ok := s.ensureTripRole(w, r, tripID, userID, "owner", "editor"); !ok {
		return
	}
	var req locationRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	if !validateLocationRequest(w, req) {
		return
	}
	location, err := s.store.UpdateLocation(r.Context(), tripID, locationID, strings.TrimSpace(req.Name), req.Description, req.Lat, req.Lng, normalizeCategory(req.Category))
	if err != nil {
		writeError(w, http.StatusNotFound, "location not found or cannot be updated")
		return
	}
	s.hub.Broadcast(tripID, models.WSEvent{Type: "LOCATION_UPDATED", Payload: location})
	writeJSON(w, http.StatusOK, location)
}

func (s *Server) handleDeleteLocation(w http.ResponseWriter, r *http.Request) {
	tripID := chi.URLParam(r, "trip_id")
	locationID := chi.URLParam(r, "location_id")
	userID, _ := appmiddleware.UserIDFromContext(r.Context())
	if _, ok := s.ensureTripRole(w, r, tripID, userID, "owner", "editor"); !ok {
		return
	}
	if err := s.store.DeleteLocation(r.Context(), tripID, locationID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete location")
		return
	}
	s.hub.Broadcast(tripID, models.WSEvent{Type: "LOCATION_DELETED", Payload: map[string]string{"id": locationID}})
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func validateLocationRequest(w http.ResponseWriter, req locationRequest) bool {
	if strings.TrimSpace(req.Name) == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return false
	}
	if req.Lat < -90 || req.Lat > 90 || req.Lng < -180 || req.Lng > 180 {
		writeError(w, http.StatusBadRequest, "invalid coordinates")
		return false
	}
	return true
}

func normalizeCategory(category string) string {
	category = strings.TrimSpace(category)
	if category == "" {
		return "other"
	}
	return category
}
