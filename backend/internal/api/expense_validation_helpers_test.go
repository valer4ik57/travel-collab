package api

import (
	"testing"

	"travel-collab/backend/internal/models"
)

func TestMoneyStringAndCentsHelpers(t *testing.T) {
	if got := toCents(10.005); got != 1001 {
		t.Fatalf("toCents(10.005) = %d, want 1001", got)
	}
	if got := centsToAmount(1001); got != 10.01 {
		t.Fatalf("centsToAmount(1001) = %.2f, want 10.01", got)
	}
	if got := moneyString(1001); got != "10.01" {
		t.Fatalf("moneyString(1001) = %q, want 10.01", got)
	}
}

func TestNormalizeOptionalID(t *testing.T) {
	value := "  route-1  "
	got := normalizeOptionalID(&value)
	if got == nil || *got != "route-1" {
		t.Fatalf("expected trimmed id, got %#v", got)
	}

	empty := "   "
	if got := normalizeOptionalID(&empty); got != nil {
		t.Fatalf("expected empty optional id to normalize to nil, got %#v", *got)
	}
	if got := normalizeOptionalID(nil); got != nil {
		t.Fatalf("expected nil optional id to stay nil")
	}
}

func TestTotalsIgnoreFloatingPointNoiseByUsingCents(t *testing.T) {
	payments := []models.ExpensePayment{{UserID: "u1", Amount: 0.1 + 0.2}, {UserID: "u2", Amount: 0.3}}
	shares := []models.ExpenseShare{{UserID: "u1", Amount: 0.2}, {UserID: "u2", Amount: 0.4}}
	if got := paymentsTotalCents(payments); got != 60 {
		t.Fatalf("paymentsTotalCents = %d, want 60", got)
	}
	if got := sharesTotalCents(shares); got != 60 {
		t.Fatalf("sharesTotalCents = %d, want 60", got)
	}
}
