package api

import (
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	appmiddleware "travel-collab/backend/internal/middleware"
	"travel-collab/backend/internal/models"
)

type createExpenseRequest struct {
	Description string                  `json:"description"`
	Amount      float64                 `json:"amount"`
	Currency    string                  `json:"currency"`
	PaidBy      string                  `json:"paid_by"`
	SplitWith   []string                `json:"split_with"`
	Payments    []expensePaymentRequest `json:"payments"`
	Shares      []expenseShareRequest   `json:"shares"`
	SplitMode   string                  `json:"split_mode"`
	RouteID     *string                 `json:"route_id"`
	LocationID  *string                 `json:"location_id"`
	ExpenseAt   *time.Time              `json:"expense_at"`
}

type expensePaymentRequest struct {
	UserID string  `json:"user_id"`
	Amount float64 `json:"amount"`
}

type expenseShareRequest struct {
	UserID string  `json:"user_id"`
	Amount float64 `json:"amount"`
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
	normalized, ok := s.normalizeAndValidateExpense(w, r, tripID, req)
	if !ok {
		return
	}
	expense, err := s.store.CreateExpense(r.Context(), tripID, normalized.Description, normalized.Amount, normalized.Currency, normalized.Payments, normalized.Shares, normalized.SplitMode, normalized.RouteID, normalized.LocationID, normalized.ExpenseAt)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create expense")
		return
	}
	summary, _ := s.store.GetExpensesSummary(r.Context(), tripID)
	s.hub.Broadcast(tripID, models.WSEvent{Type: "EXPENSE_ADDED", Payload: map[string]interface{}{"expense": expense, "summary": summary}})
	writeJSON(w, http.StatusCreated, expense)
}

type normalizedExpense struct {
	Description string
	Amount      float64
	Currency    string
	Payments    []models.ExpensePayment
	Shares      []models.ExpenseShare
	SplitMode   string
	RouteID     *string
	LocationID  *string
	ExpenseAt   *time.Time
}

func (s *Server) normalizeAndValidateExpense(w http.ResponseWriter, r *http.Request, tripID string, req createExpenseRequest) (normalizedExpense, bool) {
	result := normalizedExpense{}
	result.Description = strings.TrimSpace(req.Description)
	result.Amount = roundAmount(req.Amount)
	result.Currency = strings.ToUpper(strings.TrimSpace(req.Currency))
	if result.Currency == "" {
		result.Currency = "RUB"
	}
	result.SplitMode = strings.ToLower(strings.TrimSpace(req.SplitMode))
	if result.SplitMode == "" {
		result.SplitMode = "equal"
	}
	if result.SplitMode != "equal" && result.SplitMode != "weights" && result.SplitMode != "manual" {
		writeError(w, http.StatusBadRequest, "Способ деления расхода должен быть: поровну, по долям или ручными суммами")
		return result, false
	}
	if result.Description == "" {
		writeError(w, http.StatusBadRequest, "Введите описание расхода")
		return result, false
	}
	if toCents(result.Amount) <= 0 {
		writeError(w, http.StatusBadRequest, "Сумма расхода должна быть больше нуля")
		return result, false
	}

	result.RouteID = normalizeOptionalID(req.RouteID)
	result.LocationID = normalizeOptionalID(req.LocationID)
	result.ExpenseAt = req.ExpenseAt

	if result.LocationID != nil {
		location, err := s.store.GetLocation(r.Context(), tripID, *result.LocationID)
		if err != nil {
			writeError(w, http.StatusBadRequest, "Выбранная точка не относится к этой поездке")
			return result, false
		}
		if result.RouteID == nil && location.RouteID != nil {
			result.RouteID = location.RouteID
		}
		if result.RouteID != nil && location.RouteID != nil && *result.RouteID != *location.RouteID {
			writeError(w, http.StatusBadRequest, "Выбранная точка относится к другому маршруту")
			return result, false
		}
		if result.ExpenseAt == nil && location.VisitAt != nil {
			result.ExpenseAt = location.VisitAt
		}
	}
	if result.RouteID != nil {
		ok, err := s.store.IsRouteInTrip(r.Context(), tripID, *result.RouteID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to validate expense route")
			return result, false
		}
		if !ok {
			writeError(w, http.StatusBadRequest, "Выбранный маршрут не относится к этой поездке")
			return result, false
		}
	}

	payments := make([]models.ExpensePayment, 0)
	for _, payment := range req.Payments {
		uid := strings.TrimSpace(payment.UserID)
		amount := roundAmount(payment.Amount)
		if uid == "" || toCents(amount) <= 0 {
			continue
		}
		payments = append(payments, models.ExpensePayment{UserID: uid, Amount: amount})
	}
	if len(payments) == 0 && strings.TrimSpace(req.PaidBy) != "" {
		payments = append(payments, models.ExpensePayment{UserID: strings.TrimSpace(req.PaidBy), Amount: result.Amount})
	}
	if len(payments) == 0 {
		writeError(w, http.StatusBadRequest, "Укажите, кто и сколько оплатил")
		return result, false
	}

	shares := make([]models.ExpenseShare, 0)
	for _, share := range req.Shares {
		uid := strings.TrimSpace(share.UserID)
		amount := roundAmount(share.Amount)
		if uid == "" || toCents(amount) <= 0 {
			continue
		}
		shares = append(shares, models.ExpenseShare{UserID: uid, Amount: amount})
	}
	if len(shares) == 0 && len(req.SplitWith) > 0 {
		shares = equalShares(result.Amount, req.SplitWith)
	}
	if len(shares) == 0 {
		writeError(w, http.StatusBadRequest, "Укажите, между кем делится расход")
		return result, false
	}

	if hasDuplicatePayments(payments) {
		writeError(w, http.StatusBadRequest, "Один участник не должен повторяться в блоке оплат")
		return result, false
	}
	if hasDuplicateShares(shares) {
		writeError(w, http.StatusBadRequest, "Один участник не должен повторяться в блоке долей")
		return result, false
	}

	paymentIDs := make([]string, 0, len(payments))
	for _, payment := range payments {
		paymentIDs = append(paymentIDs, payment.UserID)
	}
	shareIDs := make([]string, 0, len(shares))
	for _, share := range shares {
		shareIDs = append(shareIDs, share.UserID)
	}
	allUserIDs := append(paymentIDs, shareIDs...)
	if err := s.store.EnsureAllUsersAreTripMembers(r.Context(), tripID, allUserIDs); err != nil {
		writeError(w, http.StatusBadRequest, "Плательщики и участники деления должны быть участниками поездки")
		return result, false
	}

	amountCents := toCents(result.Amount)
	paymentsCents := paymentsTotalCents(payments)
	sharesCents := sharesTotalCents(shares)
	if paymentsCents != amountCents {
		writeErrorDetails(w, http.StatusBadRequest, "Сумма оплат не совпадает с общей суммой расхода", map[string]interface{}{
			"total_amount": moneyString(amountCents),
			"payments_sum": moneyString(paymentsCents),
			"difference":   moneyString(absInt64(amountCents - paymentsCents)),
			"currency":     result.Currency,
		})
		return result, false
	}
	if sharesCents != amountCents {
		writeErrorDetails(w, http.StatusBadRequest, "Сумма долей не совпадает с общей суммой расхода", map[string]interface{}{
			"total_amount": moneyString(amountCents),
			"shares_sum":   moneyString(sharesCents),
			"difference":   moneyString(absInt64(amountCents - sharesCents)),
			"currency":     result.Currency,
		})
		return result, false
	}

	result.Payments = payments
	result.Shares = shares
	return result, true
}

func equalShares(amount float64, userIDs []string) []models.ExpenseShare {
	ids := make([]string, 0, len(userIDs))
	seen := map[string]bool{}
	for _, id := range userIDs {
		trimmed := strings.TrimSpace(id)
		if trimmed == "" || seen[trimmed] {
			continue
		}
		seen[trimmed] = true
		ids = append(ids, trimmed)
	}
	if len(ids) == 0 {
		return nil
	}
	totalCents := toCents(amount)
	base := totalCents / int64(len(ids))
	remainder := totalCents % int64(len(ids))
	shares := make([]models.ExpenseShare, 0, len(ids))
	for index, id := range ids {
		cents := base
		if int64(index) < remainder {
			cents++
		}
		shares = append(shares, models.ExpenseShare{UserID: id, Amount: centsToAmount(cents)})
	}
	return shares
}

func hasDuplicatePayments(items []models.ExpensePayment) bool {
	seen := map[string]bool{}
	for _, item := range items {
		if seen[item.UserID] {
			return true
		}
		seen[item.UserID] = true
	}
	return false
}

func hasDuplicateShares(items []models.ExpenseShare) bool {
	seen := map[string]bool{}
	for _, item := range items {
		if seen[item.UserID] {
			return true
		}
		seen[item.UserID] = true
	}
	return false
}

func paymentsTotalCents(items []models.ExpensePayment) int64 {
	var total int64
	for _, item := range items {
		total += toCents(item.Amount)
	}
	return total
}

func sharesTotalCents(items []models.ExpenseShare) int64 {
	var total int64
	for _, item := range items {
		total += toCents(item.Amount)
	}
	return total
}

func roundAmount(value float64) float64 {
	return math.Round(value*100) / 100
}

func toCents(value float64) int64 {
	return int64(math.Round(value * 100))
}

func centsToAmount(cents int64) float64 {
	return float64(cents) / 100
}

func absInt64(value int64) int64 {
	if value < 0 {
		return -value
	}
	return value
}

func moneyString(cents int64) string {
	return fmt.Sprintf("%.2f", centsToAmount(cents))
}

func normalizeOptionalID(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
