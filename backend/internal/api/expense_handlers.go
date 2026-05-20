package api

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	appmiddleware "travel-collab/backend/internal/middleware"
	"travel-collab/backend/internal/models"
)

type createExpenseRequest struct {
	Description string   `json:"description"`
	Amount      float64  `json:"amount"`
	Currency    string   `json:"currency"`
	PaidBy      string   `json:"paid_by"`
	SplitWith   []string `json:"split_with"`
}

func (s *Server) handleListExpenses(w http.ResponseWriter, r *http.Request) {
	tripID := chi.URLParam(r, "trip_id")
	userID, _ := appmiddleware.UserIDFromContext(r.Context())
	if !s.ensureTripMember(w, r, tripID, userID) {
		return
	}
	summary, err := s.store.GetExpensesSummary(r.Context(), tripID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list expenses")
		return
	}
	writeJSON(w, http.StatusOK, summary)
}

func (s *Server) handleCreateExpense(w http.ResponseWriter, r *http.Request) {
	tripID := chi.URLParam(r, "trip_id")
	userID, _ := appmiddleware.UserIDFromContext(r.Context())
	if _, ok := s.ensureTripRole(w, r, tripID, userID, "owner", "editor"); !ok {
		return
	}
	var req createExpenseRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	req.Description = strings.TrimSpace(req.Description)
	req.Currency = strings.ToUpper(strings.TrimSpace(req.Currency))
	if req.Currency == "" {
		req.Currency = "RUB"
	}
	if req.Description == "" || req.Amount <= 0 || req.PaidBy == "" || len(req.SplitWith) == 0 {
		writeError(w, http.StatusBadRequest, "description, amount > 0, paid_by and split_with are required")
		return
	}
	userIDs := append([]string{req.PaidBy}, req.SplitWith...)
	if err := s.store.EnsureAllUsersAreTripMembers(r.Context(), tripID, userIDs); err != nil {
		writeError(w, http.StatusBadRequest, "paid_by and split_with must be trip members")
		return
	}
	expense, err := s.store.CreateExpense(r.Context(), tripID, req.Description, req.Amount, req.Currency, req.PaidBy, req.SplitWith)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create expense")
		return
	}
	summary, _ := s.store.GetExpensesSummary(r.Context(), tripID)
	s.hub.Broadcast(tripID, models.WSEvent{Type: "EXPENSE_ADDED", Payload: map[string]interface{}{"expense": expense, "summary": summary}})
	writeJSON(w, http.StatusCreated, expense)
}
