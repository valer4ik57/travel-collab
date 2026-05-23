package repository

import (
	"context"
	"encoding/json"
	"math"
	"sort"

	"travel-collab/backend/internal/models"
)

func (s *Store) CreateExpense(ctx context.Context, tripID, description string, amount float64, currency, paidBy string, splitWith []string) (models.Expense, error) {
	bytes, err := json.Marshal(splitWith)
	if err != nil {
		return models.Expense{}, err
	}
	var e models.Expense
	var splitRaw []byte
	err = s.DB.QueryRow(ctx, `
		INSERT INTO expenses (trip_id, description, amount, currency, paid_by, split_with)
		VALUES ($1, $2, $3, $4, $5, $6::jsonb)
		RETURNING id, trip_id, description, amount::float8, currency, paid_by, split_with, created_at
	`, tripID, description, amount, currency, paidBy, string(bytes)).Scan(&e.ID, &e.TripID, &e.Description, &e.Amount, &e.Currency, &e.PaidBy, &splitRaw, &e.CreatedAt)
	if err != nil {
		return e, err
	}
	_ = json.Unmarshal(splitRaw, &e.SplitWith)
	return e, nil
}

func (s *Store) ListExpenses(ctx context.Context, tripID string) ([]models.Expense, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT e.id, e.trip_id, e.description, e.amount::float8, e.currency, e.paid_by, u.display_name, e.split_with, e.created_at
		FROM expenses e
		JOIN users u ON u.id = e.paid_by
		WHERE e.trip_id = $1
		ORDER BY e.created_at DESC
	`, tripID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	expenses := make([]models.Expense, 0)
	for rows.Next() {
		var e models.Expense
		var splitRaw []byte
		if err := rows.Scan(&e.ID, &e.TripID, &e.Description, &e.Amount, &e.Currency, &e.PaidBy, &e.PaidByName, &splitRaw, &e.CreatedAt); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(splitRaw, &e.SplitWith)
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
	for _, e := range expenses {
		currency = e.Currency
		totalAmount += e.Amount
		if _, ok := paidByUser[e.PaidBy]; !ok {
			paidByUser[e.PaidBy] = 0
			shareByUser[e.PaidBy] = 0
			nameByUser[e.PaidBy] = e.PaidByName
		}
		paidByUser[e.PaidBy] += e.Amount
		if len(e.SplitWith) == 0 {
			continue
		}
		share := e.Amount / float64(len(e.SplitWith))
		for _, uid := range e.SplitWith {
			if _, ok := shareByUser[uid]; !ok {
				paidByUser[uid] = 0
				shareByUser[uid] = 0
				nameByUser[uid] = "Участник вне поездки"
			}
			shareByUser[uid] += share
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

	settlements := buildSettlements(balances, currency)
	return models.ExpensesSummary{Expenses: expenses, Balances: balances, Participants: participants, Settlements: settlements, TotalAmount: round2(totalAmount), Currency: currency}, nil
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
