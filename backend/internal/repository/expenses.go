package repository

import (
	"context"
	"encoding/json"
	"math"
	"sort"

	"github.com/jackc/pgx/v5"

	"travel-collab/backend/internal/models"
)

func (s *Store) CreateExpense(ctx context.Context, tripID, description string, amount float64, currency string, payments []models.ExpensePayment, shares []models.ExpenseShare, splitMode string, routeID *string, locationID *string, expenseAt any) (models.Expense, error) {
	paymentsBytes, err := json.Marshal(payments)
	if err != nil {
		return models.Expense{}, err
	}
	sharesBytes, err := json.Marshal(shares)
	if err != nil {
		return models.Expense{}, err
	}
	splitWith := make([]string, 0, len(shares))
	for _, share := range shares {
		splitWith = append(splitWith, share.UserID)
	}
	splitBytes, err := json.Marshal(splitWith)
	if err != nil {
		return models.Expense{}, err
	}
	paidBy := ""
	if len(payments) > 0 {
		paidBy = payments[0].UserID
	}
	var routeParam any
	if routeID != nil {
		routeParam = *routeID
	}
	var locationParam any
	if locationID != nil {
		locationParam = *locationID
	}

	var e models.Expense
	err = s.DB.QueryRow(ctx, `
		INSERT INTO expenses (trip_id, description, amount, currency, paid_by, split_with, location_id, route_id, expense_at, payments, shares, split_mode)
		VALUES ($1, $2, $3, $4, $5, $6::jsonb, $7, $8, $9, $10::jsonb, $11::jsonb, $12)
		RETURNING id
	`, tripID, description, amount, currency, paidBy, string(splitBytes), locationParam, routeParam, expenseAt, string(paymentsBytes), string(sharesBytes), splitMode).Scan(&e.ID)
	if err != nil {
		return e, err
	}
	expenses, err := s.ListExpenses(ctx, tripID)
	if err != nil {
		return e, err
	}
	for _, item := range expenses {
		if item.ID == e.ID {
			return item, nil
		}
	}
	return e, nil
}

func (s *Store) UpdateExpense(ctx context.Context, tripID, expenseID, description string, amount float64, currency string, payments []models.ExpensePayment, shares []models.ExpenseShare, splitMode string, routeID *string, locationID *string, expenseAt any) (models.Expense, error) {
	paymentsBytes, err := json.Marshal(payments)
	if err != nil {
		return models.Expense{}, err
	}
	sharesBytes, err := json.Marshal(shares)
	if err != nil {
		return models.Expense{}, err
	}
	splitWith := make([]string, 0, len(shares))
	for _, share := range shares {
		splitWith = append(splitWith, share.UserID)
	}
	splitBytes, err := json.Marshal(splitWith)
	if err != nil {
		return models.Expense{}, err
	}
	paidBy := ""
	if len(payments) > 0 {
		paidBy = payments[0].UserID
	}
	var routeParam any
	if routeID != nil {
		routeParam = *routeID
	}
	var locationParam any
	if locationID != nil {
		locationParam = *locationID
	}

	var updatedID string
	err = s.DB.QueryRow(ctx, `
		UPDATE expenses
		SET description = $3,
		    amount = $4,
		    currency = $5,
		    paid_by = $6,
		    split_with = $7::jsonb,
		    location_id = $8,
		    route_id = $9,
		    expense_at = $10,
		    payments = $11::jsonb,
		    shares = $12::jsonb,
		    split_mode = $13
		WHERE trip_id = $1 AND id = $2
		RETURNING id
	`, tripID, expenseID, description, amount, currency, paidBy, string(splitBytes), locationParam, routeParam, expenseAt, string(paymentsBytes), string(sharesBytes), splitMode).Scan(&updatedID)
	if err != nil {
		return models.Expense{}, err
	}

	expenses, err := s.ListExpenses(ctx, tripID)
	if err != nil {
		return models.Expense{}, err
	}
	for _, item := range expenses {
		if item.ID == updatedID {
			return item, nil
		}
	}
	return models.Expense{}, pgx.ErrNoRows
}

func (s *Store) DeleteExpense(ctx context.Context, tripID, expenseID string) error {
	ct, err := s.DB.Exec(ctx, `
		DELETE FROM expenses
		WHERE trip_id = $1 AND id = $2
	`, tripID, expenseID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (s *Store) ListExpenses(ctx context.Context, tripID string) ([]models.Expense, error) {
	nameByUser := map[string]string{}
	if members, err := s.ListTripMembers(ctx, tripID); err == nil {
		for _, member := range members {
			nameByUser[member.UserID] = member.DisplayName
		}
	}

	rows, err := s.DB.Query(ctx, `
		SELECT e.id, e.trip_id, e.description, e.amount::float8, e.currency, e.paid_by, u.display_name,
		       e.split_with, e.payments, e.shares, e.split_mode,
		       e.route_id::text, r.title, r.route_date,
		       e.location_id::text, l.name, e.expense_at, e.created_at
		FROM expenses e
		JOIN users u ON u.id = e.paid_by
		LEFT JOIN trip_routes r ON r.id = e.route_id
		LEFT JOIN locations l ON l.id = e.location_id
		WHERE e.trip_id = $1
		ORDER BY
			CASE WHEN e.expense_at IS NULL THEN 1 ELSE 0 END,
			e.expense_at DESC,
			e.created_at DESC
	`, tripID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	expenses := make([]models.Expense, 0)
	for rows.Next() {
		var e models.Expense
		var splitRaw, paymentsRaw, sharesRaw []byte
		if err := rows.Scan(&e.ID, &e.TripID, &e.Description, &e.Amount, &e.Currency, &e.PaidBy, &e.PaidByName, &splitRaw, &paymentsRaw, &sharesRaw, &e.SplitMode, &e.RouteID, &e.RouteTitle, &e.RouteDate, &e.LocationID, &e.LocationName, &e.ExpenseAt, &e.CreatedAt); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(splitRaw, &e.SplitWith)
		_ = json.Unmarshal(paymentsRaw, &e.Payments)
		_ = json.Unmarshal(sharesRaw, &e.Shares)
		enrichExpenseFinancials(&e, nameByUser)
		expenses = append(expenses, e)
	}
	return expenses, rows.Err()
}

func (s *Store) GetExpensesSummary(ctx context.Context, tripID string) (models.ExpensesSummary, error) {
	members, err := s.ListTripMembers(ctx, tripID)
	if err != nil {
		return models.ExpensesSummary{}, err
	}
	expenses, err := s.ListExpenses(ctx, tripID)
	if err != nil {
		return models.ExpensesSummary{}, err
	}

	currency := "RUB"
	paidByUser := make(map[string]float64)
	shareByUser := make(map[string]float64)
	nameByUser := make(map[string]string)
	for _, m := range members {
		paidByUser[m.UserID] = 0
		shareByUser[m.UserID] = 0
		nameByUser[m.UserID] = m.DisplayName
	}

	totalAmount := 0.0
	totalPaid := 0.0
	totalShares := 0.0
	routeTotals := make(map[string]*models.RouteExpenseSummary)
	for _, e := range expenses {
		currency = e.Currency
		totalAmount += e.Amount
		key := "__general__"
		title := "Общие расходы поездки"
		var routeID *string
		if e.RouteID != nil {
			key = *e.RouteID
			routeID = e.RouteID
			if e.RouteTitle != nil && *e.RouteTitle != "" {
				title = *e.RouteTitle
			}
		}
		if _, ok := routeTotals[key]; !ok {
			routeTotals[key] = &models.RouteExpenseSummary{RouteID: routeID, RouteTitle: title, RouteDate: e.RouteDate, Currency: currency}
		}
		routeTotals[key].TotalAmount += e.Amount
		routeTotals[key].ExpensesCount++

		for _, payment := range effectivePayments(e) {
			if _, ok := paidByUser[payment.UserID]; !ok {
				paidByUser[payment.UserID] = 0
				shareByUser[payment.UserID] = 0
				nameByUser[payment.UserID] = displayName(payment.DisplayName, "Участник вне поездки")
			}
			paidByUser[payment.UserID] += payment.Amount
			totalPaid += payment.Amount
		}
		for _, share := range effectiveShares(e) {
			if _, ok := shareByUser[share.UserID]; !ok {
				paidByUser[share.UserID] = 0
				shareByUser[share.UserID] = 0
				nameByUser[share.UserID] = displayName(share.DisplayName, "Участник вне поездки")
			}
			shareByUser[share.UserID] += share.Amount
			totalShares += share.Amount
		}
	}

	balances := make([]models.Balance, 0, len(paidByUser))
	participants := make([]models.ExpenseParticipantSummary, 0, len(paidByUser))
	for uid, paid := range paidByUser {
		share := shareByUser[uid]
		net := paid - share
		balances = append(balances, models.Balance{
			UserID:      uid,
			DisplayName: nameByUser[uid],
			Amount:      round2(net),
			Currency:    currency,
		})
		participants = append(participants, models.ExpenseParticipantSummary{
			UserID:      uid,
			DisplayName: nameByUser[uid],
			PaidTotal:   round2(paid),
			ShareTotal:  round2(share),
			NetBalance:  round2(net),
			Currency:    currency,
		})
	}
	sort.Slice(balances, func(i, j int) bool { return balances[i].DisplayName < balances[j].DisplayName })
	sort.Slice(participants, func(i, j int) bool { return participants[i].DisplayName < participants[j].DisplayName })

	routeSummaries := make([]models.RouteExpenseSummary, 0, len(routeTotals))
	for _, item := range routeTotals {
		item.TotalAmount = round2(item.TotalAmount)
		routeSummaries = append(routeSummaries, *item)
	}
	sort.Slice(routeSummaries, func(i, j int) bool {
		if routeSummaries[i].RouteDate == nil && routeSummaries[j].RouteDate != nil {
			return false
		}
		if routeSummaries[i].RouteDate != nil && routeSummaries[j].RouteDate == nil {
			return true
		}
		if routeSummaries[i].RouteDate != nil && routeSummaries[j].RouteDate != nil && !routeSummaries[i].RouteDate.Equal(*routeSummaries[j].RouteDate) {
			return routeSummaries[i].RouteDate.Before(*routeSummaries[j].RouteDate)
		}
		return routeSummaries[i].RouteTitle < routeSummaries[j].RouteTitle
	})

	settlements := buildSettlements(balances, currency)
	return models.ExpensesSummary{
		Expenses:       expenses,
		Balances:       balances,
		Participants:   participants,
		Settlements:    settlements,
		RouteSummaries: routeSummaries,
		TotalAmount:    round2(totalAmount),
		TotalPaid:      round2(totalPaid),
		TotalShares:    round2(totalShares),
		Currency:       currency,
	}, nil
}

func enrichExpenseFinancials(e *models.Expense, names map[string]string) {
	if len(e.Payments) == 0 && e.PaidBy != "" {
		e.Payments = []models.ExpensePayment{{UserID: e.PaidBy, DisplayName: displayName(e.PaidByName, names[e.PaidBy]), Amount: e.Amount}}
	}
	if len(e.Shares) == 0 && len(e.SplitWith) > 0 {
		share := round2(e.Amount / float64(len(e.SplitWith)))
		remainder := round2(e.Amount - share*float64(len(e.SplitWith)))
		e.Shares = make([]models.ExpenseShare, 0, len(e.SplitWith))
		for index, uid := range e.SplitWith {
			amount := share
			if index == len(e.SplitWith)-1 {
				amount = round2(amount + remainder)
			}
			e.Shares = append(e.Shares, models.ExpenseShare{UserID: uid, DisplayName: displayName(names[uid], "Участник вне поездки"), Amount: amount})
		}
	}
	for index := range e.Payments {
		e.Payments[index].DisplayName = displayName(e.Payments[index].DisplayName, names[e.Payments[index].UserID])
	}
	for index := range e.Shares {
		e.Shares[index].DisplayName = displayName(e.Shares[index].DisplayName, names[e.Shares[index].UserID])
	}
	if e.SplitMode == "" {
		e.SplitMode = "equal"
	}
}

func effectivePayments(e models.Expense) []models.ExpensePayment {
	if len(e.Payments) > 0 {
		return e.Payments
	}
	if e.PaidBy == "" {
		return nil
	}
	return []models.ExpensePayment{{UserID: e.PaidBy, DisplayName: e.PaidByName, Amount: e.Amount}}
}

func effectiveShares(e models.Expense) []models.ExpenseShare {
	if len(e.Shares) > 0 {
		return e.Shares
	}
	if len(e.SplitWith) == 0 {
		return nil
	}
	share := round2(e.Amount / float64(len(e.SplitWith)))
	remainder := round2(e.Amount - share*float64(len(e.SplitWith)))
	result := make([]models.ExpenseShare, 0, len(e.SplitWith))
	for index, uid := range e.SplitWith {
		amount := share
		if index == len(e.SplitWith)-1 {
			amount = round2(amount + remainder)
		}
		result = append(result, models.ExpenseShare{UserID: uid, Amount: amount})
	}
	return result
}

func displayName(value string, fallback string) string {
	if value != "" {
		return value
	}
	if fallback != "" {
		return fallback
	}
	return "Участник"
}

func buildSettlements(balances []models.Balance, currency string) []models.Settlement {
	type person struct {
		id     string
		name   string
		amount float64
	}
	var debtors, creditors []person
	for _, b := range balances {
		if b.Amount < -0.005 {
			debtors = append(debtors, person{id: b.UserID, name: b.DisplayName, amount: -b.Amount})
		} else if b.Amount > 0.005 {
			creditors = append(creditors, person{id: b.UserID, name: b.DisplayName, amount: b.Amount})
		}
	}
	result := make([]models.Settlement, 0)
	i, j := 0, 0
	for i < len(debtors) && j < len(creditors) {
		amount := math.Min(debtors[i].amount, creditors[j].amount)
		if amount > 0.005 {
			result = append(result, models.Settlement{
				FromUserID: debtors[i].id,
				FromName:   debtors[i].name,
				ToUserID:   creditors[j].id,
				ToName:     creditors[j].name,
				Amount:     round2(amount),
				Currency:   currency,
			})
		}
		debtors[i].amount -= amount
		creditors[j].amount -= amount
		if debtors[i].amount <= 0.005 {
			i++
		}
		if creditors[j].amount <= 0.005 {
			j++
		}
	}
	return result
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}
