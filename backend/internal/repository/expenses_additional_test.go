package repository

import (
	"testing"

	"travel-collab/backend/internal/models"
)

func TestBuildSettlementsIgnoresNearZeroBalances(t *testing.T) {
	balances := []models.Balance{
		{UserID: "alex", DisplayName: "Алексей", Amount: 0.004, Currency: "RUB"},
		{UserID: "ivan", DisplayName: "Иван", Amount: -0.004, Currency: "RUB"},
	}
	settlements := buildSettlements(balances, "RUB")
	if len(settlements) != 0 {
		t.Fatalf("expected no settlements for near-zero balances, got %#v", settlements)
	}
}

func TestEffectiveSharesFallbackAssignsRemainderToLastParticipant(t *testing.T) {
	expense := models.Expense{Amount: 100, SplitWith: []string{"alex", "ivan", "maria"}}
	shares := effectiveShares(expense)
	if len(shares) != 3 {
		t.Fatalf("expected 3 shares, got %d", len(shares))
	}
	if shares[0].Amount != 33.33 || shares[1].Amount != 33.33 || shares[2].Amount != 33.34 {
		t.Fatalf("unexpected shares: %#v", shares)
	}
}

func TestEnrichExpenseFinancialsSetsDisplayNamesAndSplitMode(t *testing.T) {
	expense := models.Expense{
		Amount:    90,
		PaidBy:    "alex",
		SplitWith: []string{"alex", "ivan"},
	}
	names := map[string]string{"alex": "Алексей", "ivan": "Иван"}
	enrichExpenseFinancials(&expense, names)

	if expense.SplitMode != "equal" {
		t.Fatalf("expected empty split mode to become equal, got %q", expense.SplitMode)
	}
	if len(expense.Payments) != 1 || expense.Payments[0].DisplayName != "Алексей" {
		t.Fatalf("unexpected payments after enrich: %#v", expense.Payments)
	}
	if len(expense.Shares) != 2 || expense.Shares[1].DisplayName != "Иван" {
		t.Fatalf("unexpected shares after enrich: %#v", expense.Shares)
	}
}
