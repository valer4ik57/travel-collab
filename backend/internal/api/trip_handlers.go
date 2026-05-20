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

type createTripRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
	StartDate   *string `json:"start_date"`
	EndDate     *string `json:"end_date"`
}

type joinTripRequest struct {
	InviteCode string `json:"invite_code"`
}

func (s *Server) handleListTrips(w http.ResponseWriter, r *http.Request) {
	userID, _ := appmiddleware.UserIDFromContext(r.Context())
	trips, err := s.store.ListTripsForUser(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list trips")
		return
	}
	writeJSON(w, http.StatusOK, trips)
}

func (s *Server) handleCreateTrip(w http.ResponseWriter, r *http.Request) {
	userID, _ := appmiddleware.UserIDFromContext(r.Context())
	var req createTripRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	var startDate, endDate *time.Time
	if req.StartDate != nil && strings.TrimSpace(*req.StartDate) != "" {
		parsed, err := time.Parse("2006-01-02", *req.StartDate)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid start_date format; use YYYY-MM-DD")
			return
		}
		startDate = &parsed
	}
	if req.EndDate != nil && strings.TrimSpace(*req.EndDate) != "" {
		parsed, err := time.Parse("2006-01-02", *req.EndDate)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid end_date format; use YYYY-MM-DD")
			return
		}
		endDate = &parsed
	}
	trip, err := s.store.CreateTrip(r.Context(), userID, name, req.Description, startDate, endDate)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create trip")
		return
	}
	writeJSON(w, http.StatusCreated, trip)
}

func (s *Server) handleGetTrip(w http.ResponseWriter, r *http.Request) {
	tripID := chi.URLParam(r, "trip_id")
	userID, _ := appmiddleware.UserIDFromContext(r.Context())
	if !s.ensureTripMember(w, r, tripID, userID) {
		return
	}
	details, err := s.store.GetTripDetails(r.Context(), tripID)
	if err != nil {
		if err == pgx.ErrNoRows {
			writeError(w, http.StatusNotFound, "trip not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to get trip")
		return
	}
	writeJSON(w, http.StatusOK, details)
}

func (s *Server) handleJoinTrip(w http.ResponseWriter, r *http.Request) {
	userID, _ := appmiddleware.UserIDFromContext(r.Context())
	var req joinTripRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	inviteCode := strings.ToUpper(strings.TrimSpace(req.InviteCode))
	if inviteCode == "" {
		writeError(w, http.StatusBadRequest, "invite_code is required")
		return
	}
	trip, member, created, err := s.store.JoinTripByInviteCode(r.Context(), userID, inviteCode)
	if err != nil {
		if err == pgx.ErrNoRows {
			writeError(w, http.StatusNotFound, "invite code not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to join trip")
		return
	}
	if created {
		s.hub.Broadcast(trip.ID, models.WSEvent{Type: "MEMBER_JOINED", Payload: member})
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"trip": trip, "member": member, "already_member": !created})
}

func (s *Server) ensureTripMember(w http.ResponseWriter, r *http.Request, tripID, userID string) bool {
	if tripID == "" || userID == "" {
		writeError(w, http.StatusBadRequest, "trip_id is required")
		return false
	}
	ok, err := s.store.IsTripMember(r.Context(), tripID, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to check trip access")
		return false
	}
	if !ok {
		writeError(w, http.StatusForbidden, "you are not a member of this trip")
		return false
	}
	return true
}
