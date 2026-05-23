package api

import (
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	appmiddleware "travel-collab/backend/internal/middleware"
	"travel-collab/backend/internal/models"
)

type locationRequest struct {
	Name        string     `json:"name"`
	Description *string    `json:"description"`
	Lat         float64    `json:"lat"`
	Lng         float64    `json:"lng"`
	Category    string     `json:"category"`
	VisitAt     *time.Time `json:"visit_at"`
	RouteID     *string    `json:"route_id"`
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
	if !s.validateLocationRequest(w, r, tripID, req) {
		return
	}
	routeID := normalizeOptionalID(req.RouteID)
	location, err := s.store.CreateLocation(r.Context(), tripID, userID, strings.TrimSpace(req.Name), cleanOptionalString(req.Description), req.Lat, req.Lng, normalizeCategory(req.Category), req.VisitAt, routeID)
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
	if !s.validateLocationRequest(w, r, tripID, req) {
		return
	}
	routeID := normalizeOptionalID(req.RouteID)
	location, err := s.store.UpdateLocation(r.Context(), tripID, locationID, strings.TrimSpace(req.Name), cleanOptionalString(req.Description), req.Lat, req.Lng, normalizeCategory(req.Category), req.VisitAt, routeID)
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

func (s *Server) validateLocationRequest(w http.ResponseWriter, r *http.Request, tripID string, req locationRequest) bool {
	if strings.TrimSpace(req.Name) == "" {
		writeError(w, http.StatusBadRequest, "Введите название точки")
		return false
	}
	if req.Lat < -90 || req.Lat > 90 || req.Lng < -180 || req.Lng > 180 {
		writeErrorDetails(w, http.StatusBadRequest, "Координаты точки некорректны", map[string]interface{}{
			"lat": req.Lat,
			"lng": req.Lng,
		})
		return false
	}
	routeID := normalizeOptionalID(req.RouteID)
	if routeID != nil {
		ok, err := s.store.IsRouteInTrip(r.Context(), tripID, *routeID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to validate route")
			return false
		}
		if !ok {
			writeError(w, http.StatusBadRequest, "Выбранный маршрут не относится к этой поездке")
			return false
		}
	}
	return true
}

func normalizeCategory(category string) string {
	switch strings.ToLower(strings.TrimSpace(category)) {
	case "sight", "food", "hotel", "transport", "other":
		return strings.ToLower(strings.TrimSpace(category))
	case "":
		return "other"
	default:
		return "other"
	}
}

func cleanOptionalString(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
