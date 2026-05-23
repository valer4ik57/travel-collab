package api

import (
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"

	appmiddleware "travel-collab/backend/internal/middleware"
	"travel-collab/backend/internal/models"
)

type routeRequest struct {
	Title     string  `json:"title"`
	RouteDate *string `json:"route_date"`
}

type reorderLocationsRequest struct {
	LocationIDs []string `json:"location_ids"`
}

func (s *Server) handleListRoutes(w http.ResponseWriter, r *http.Request) {
	tripID := chi.URLParam(r, "trip_id")
	userID, _ := appmiddleware.UserIDFromContext(r.Context())
	if !s.ensureTripMember(w, r, tripID, userID) {
		return
	}
	routes, err := s.store.ListRoutes(r.Context(), tripID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list routes")
		return
	}
	writeJSON(w, http.StatusOK, routes)
}

func (s *Server) handleCreateRoute(w http.ResponseWriter, r *http.Request) {
	tripID := chi.URLParam(r, "trip_id")
	userID, _ := appmiddleware.UserIDFromContext(r.Context())
	if _, ok := s.ensureTripRole(w, r, tripID, userID, "owner", "editor"); !ok {
		return
	}
	var req routeRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	title := strings.TrimSpace(req.Title)
	if title == "" {
		writeError(w, http.StatusBadRequest, "Введите название маршрута")
		return
	}
	routeDate, ok := parseOptionalDate(w, req.RouteDate)
	if !ok {
		return
	}
	route, err := s.store.CreateRoute(r.Context(), tripID, title, routeDate)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create route")
		return
	}
	s.hub.Broadcast(tripID, models.WSEvent{Type: "ROUTE_ADDED", Payload: route})
	writeJSON(w, http.StatusCreated, route)
}

func (s *Server) handleUpdateRoute(w http.ResponseWriter, r *http.Request) {
	tripID := chi.URLParam(r, "trip_id")
	routeID := chi.URLParam(r, "route_id")
	userID, _ := appmiddleware.UserIDFromContext(r.Context())
	if _, ok := s.ensureTripRole(w, r, tripID, userID, "owner", "editor"); !ok {
		return
	}
	var req routeRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	title := strings.TrimSpace(req.Title)
	if title == "" {
		writeError(w, http.StatusBadRequest, "Введите название маршрута")
		return
	}
	routeDate, ok := parseOptionalDate(w, req.RouteDate)
	if !ok {
		return
	}
	route, err := s.store.UpdateRoute(r.Context(), tripID, routeID, title, routeDate)
	if err != nil {
		if err == pgx.ErrNoRows {
			writeError(w, http.StatusNotFound, "Маршрут не найден")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to update route")
		return
	}
	s.hub.Broadcast(tripID, models.WSEvent{Type: "ROUTE_UPDATED", Payload: route})
	writeJSON(w, http.StatusOK, route)
}

func (s *Server) handleDeleteRoute(w http.ResponseWriter, r *http.Request) {
	tripID := chi.URLParam(r, "trip_id")
	routeID := chi.URLParam(r, "route_id")
	userID, _ := appmiddleware.UserIDFromContext(r.Context())
	if _, ok := s.ensureTripRole(w, r, tripID, userID, "owner", "editor"); !ok {
		return
	}
	locationsCount, expensesCount, err := s.store.CountRouteLinks(r.Context(), tripID, routeID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to check route links")
		return
	}
	if locationsCount > 0 || expensesCount > 0 {
		writeErrorDetails(w, http.StatusBadRequest, "Нельзя удалить маршрут, пока к нему привязаны точки или расходы", map[string]interface{}{
			"locations_count": locationsCount,
			"expenses_count":  expensesCount,
		})
		return
	}
	if err := s.store.DeleteRoute(r.Context(), tripID, routeID); err != nil {
		if err == pgx.ErrNoRows {
			writeError(w, http.StatusNotFound, "Маршрут не найден")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to delete route")
		return
	}
	s.hub.Broadcast(tripID, models.WSEvent{Type: "ROUTE_DELETED", Payload: map[string]string{"id": routeID}})
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (s *Server) handleReorderRouteLocations(w http.ResponseWriter, r *http.Request) {
	tripID := chi.URLParam(r, "trip_id")
	routeID := chi.URLParam(r, "route_id")
	userID, _ := appmiddleware.UserIDFromContext(r.Context())
	if _, ok := s.ensureTripRole(w, r, tripID, userID, "owner", "editor"); !ok {
		return
	}
	var req reorderLocationsRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	if len(req.LocationIDs) == 0 {
		writeError(w, http.StatusBadRequest, "Передайте список точек в новом порядке")
		return
	}
	if ok, err := s.store.IsRouteInTrip(r.Context(), tripID, routeID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to validate route")
		return
	} else if !ok {
		writeError(w, http.StatusNotFound, "Маршрут не найден")
		return
	}
	locations, err := s.store.ReorderRouteLocations(r.Context(), tripID, routeID, req.LocationIDs)
	if err != nil {
		if err == pgx.ErrNoRows {
			writeError(w, http.StatusBadRequest, "Одна из точек не относится к выбранному маршруту")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to reorder route locations")
		return
	}
	s.hub.Broadcast(tripID, models.WSEvent{Type: "LOCATIONS_REORDERED", Payload: locations})
	writeJSON(w, http.StatusOK, locations)
}

func parseOptionalDate(w http.ResponseWriter, raw *string) (*time.Time, bool) {
	if raw == nil || strings.TrimSpace(*raw) == "" {
		return nil, true
	}
	parsed, err := time.Parse("2006-01-02", strings.TrimSpace(*raw))
	if err != nil {
		writeError(w, http.StatusBadRequest, "Дата маршрута должна быть в формате YYYY-MM-DD")
		return nil, false
	}
	return &parsed, true
}
