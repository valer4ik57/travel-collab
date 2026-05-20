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

type updateMemberRoleRequest struct {
	Role string `json:"role"`
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

func (s *Server) handleLeaveTrip(w http.ResponseWriter, r *http.Request) {
	tripID := chi.URLParam(r, "trip_id")
	userID, _ := appmiddleware.UserIDFromContext(r.Context())
	role, ok := s.ensureTripRole(w, r, tripID, userID, "owner", "editor", "viewer")
	if !ok {
		return
	}
	if role == "owner" {
		owners, err := s.store.CountOwners(r.Context(), tripID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to check owners")
			return
		}
		if owners <= 1 {
			writeError(w, http.StatusBadRequest, "owner cannot leave while they are the only owner")
			return
		}
	}
	if err := s.store.RemoveTripMember(r.Context(), tripID, userID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to leave trip")
		return
	}
	s.hub.Broadcast(tripID, models.WSEvent{Type: "MEMBER_LEFT", Payload: map[string]string{"user_id": userID}})
	writeJSON(w, http.StatusOK, map[string]string{"status": "left"})
}

func (s *Server) handleUpdateMemberRole(w http.ResponseWriter, r *http.Request) {
	tripID := chi.URLParam(r, "trip_id")
	targetUserID := chi.URLParam(r, "user_id")
	userID, _ := appmiddleware.UserIDFromContext(r.Context())
	if _, ok := s.ensureTripRole(w, r, tripID, userID, "owner"); !ok {
		return
	}
	var req updateMemberRoleRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	role := strings.ToLower(strings.TrimSpace(req.Role))
	if role != "owner" && role != "editor" && role != "viewer" {
		writeError(w, http.StatusBadRequest, "role must be owner, editor or viewer")
		return
	}
	oldRole, err := s.store.GetTripMemberRole(r.Context(), tripID, targetUserID)
	if err != nil {
		writeError(w, http.StatusNotFound, "member not found")
		return
	}
	if oldRole == "owner" && role != "owner" {
		owners, err := s.store.CountOwners(r.Context(), tripID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to check owners")
			return
		}
		if owners <= 1 {
			writeError(w, http.StatusBadRequest, "cannot remove the last owner role")
			return
		}
	}
	member, err := s.store.UpdateTripMemberRole(r.Context(), tripID, targetUserID, role)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update member role")
		return
	}
	s.hub.Broadcast(tripID, models.WSEvent{Type: "MEMBER_UPDATED", Payload: member})
	writeJSON(w, http.StatusOK, member)
}

func (s *Server) handleRemoveMember(w http.ResponseWriter, r *http.Request) {
	tripID := chi.URLParam(r, "trip_id")
	targetUserID := chi.URLParam(r, "user_id")
	userID, _ := appmiddleware.UserIDFromContext(r.Context())
	if _, ok := s.ensureTripRole(w, r, tripID, userID, "owner"); !ok {
		return
	}
	if userID == targetUserID {
		writeError(w, http.StatusBadRequest, "use leave trip for yourself")
		return
	}
	role, err := s.store.GetTripMemberRole(r.Context(), tripID, targetUserID)
	if err != nil {
		writeError(w, http.StatusNotFound, "member not found")
		return
	}
	if role == "owner" {
		owners, err := s.store.CountOwners(r.Context(), tripID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to check owners")
			return
		}
		if owners <= 1 {
			writeError(w, http.StatusBadRequest, "cannot remove the last owner")
			return
		}
	}
	if err := s.store.RemoveTripMember(r.Context(), tripID, targetUserID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to remove member")
		return
	}
	s.hub.Broadcast(tripID, models.WSEvent{Type: "MEMBER_REMOVED", Payload: map[string]string{"user_id": targetUserID}})
	writeJSON(w, http.StatusOK, map[string]string{"status": "removed"})
}

func (s *Server) ensureTripMember(w http.ResponseWriter, r *http.Request, tripID, userID string) bool {
	_, ok := s.ensureTripRole(w, r, tripID, userID, "owner", "editor", "viewer")
	return ok
}

func (s *Server) ensureTripRole(w http.ResponseWriter, r *http.Request, tripID, userID string, allowed ...string) (string, bool) {
	if tripID == "" || userID == "" {
		writeError(w, http.StatusBadRequest, "trip_id is required")
		return "", false
	}
	role, err := s.store.GetTripMemberRole(r.Context(), tripID, userID)
	if err != nil {
		if err == pgx.ErrNoRows {
			writeError(w, http.StatusForbidden, "you are not a member of this trip")
			return "", false
		}
		writeError(w, http.StatusInternalServerError, "failed to check trip access")
		return "", false
	}
	for _, item := range allowed {
		if role == item {
			return role, true
		}
	}
	writeError(w, http.StatusForbidden, "your role does not allow this action")
	return role, false
}
