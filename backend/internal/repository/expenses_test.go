package repository

import (
	"testing"

	"travel-collab/backend/internal/models"
)

func TestBuildSettlementsSplitsOneDebtorBetweenCreditors(t *testing.T) {
	balances := []models.Balance{
		{UserID: "alex", DisplayName: "Алексей", Amount: 3000, Currency: "RUB"},
		{UserID: "maria", DisplayName: "Мария", Amount: 2000, Currency: "RUB"},
		{UserID: "ivan", DisplayName: "Иван", Amount: -5000, Currency: "RUB"},
	}

	settlements := buildSettlements(balances, "RUB")
	if len(settlements) != 2 {
		t.Fatalf("expected 2 settlements, got %d: %#v", len(settlements), settlements)
	}
	if settlements[0].FromUserID != "ivan" || settlements[0].ToUserID != "alex" || settlements[0].Amount != 3000 {
		t.Fatalf("unexpected first settlement: %#v", settlements[0])
	}
	if settlements[1].FromUserID != "ivan" || settlements[1].ToUserID != "maria" || settlements[1].Amount != 2000 {
		t.Fatalf("unexpected second settlement: %#v", settlements[1])
	}
}

func TestBuildSettlementsHandlesMultipleDebtors(t *testing.T) {
	balances := []models.Balance{
		{UserID: "alex", DisplayName: "Алексей", Amount: 500, Currency: "RUB"},
		{UserID: "maria", DisplayName: "Мария", Amount: 500, Currency: "RUB"},
		{UserID: "ivan", DisplayName: "Иван", Amount: -700, Currency: "RUB"},
		{UserID: "petr", DisplayName: "Пётр", Amount: -300, Currency: "RUB"},
	}

	settlements := buildSettlements(balances, "RUB")
	if len(settlements) != 3 {
		t.Fatalf("expected 3 settlements, got %d: %#v", len(settlements), settlements)
	}

	var total float64
	for _, settlement := range settlements {
		if settlement.Amount <= 0 {
			t.Fatalf("settlement amount should be positive: %#v", settlement)
		}
		total += settlement.Amount
	}
	if round2(total) != 1000 {
		t.Fatalf("expected total settlement amount 1000, got %.2f", total)
	}
}

func TestEffectiveSharesKeepsExpenseTotalAfterRounding(t *testing.T) {
	expense := models.Expense{Amount: 100, SplitWith: []string{"alex", "ivan", "maria"}}
	shares := effectiveShares(expense)
	if len(shares) != 3 {
		t.Fatalf("expected 3 shares, got %d", len(shares))
	}

	var total float64
	for _, share := range shares {
		total += share.Amount
	}
	if round2(total) != 100 {
		t.Fatalf("expected total 100, got %.2f", total)
	}
}

func TestEffectivePaymentsFallbackToLegacyPaidBy(t *testing.T) {
	expense := models.Expense{Amount: 1200, PaidBy: "alex", PaidByName: "Алексей"}
	payments := effectivePayments(expense)
	if len(payments) != 1 {
		t.Fatalf("expected 1 payment, got %d", len(payments))
	}
	if payments[0].UserID != "alex" || payments[0].Amount != 1200 {
		t.Fatalf("unexpected fallback payment: %#v", payments[0])
	}
}
