package api

import (
	"testing"

	"travel-collab/backend/internal/models"
)

func TestEqualSharesDistributesRemainderByCents(t *testing.T) {
	shares := equalShares(100, []string{"alex", "ivan", "maria"})
	if len(shares) != 3 {
		t.Fatalf("expected 3 shares, got %d", len(shares))
	}

	if got := sharesTotalCents(shares); got != 10000 {
		t.Fatalf("expected total 10000 cents, got %d", got)
	}

	if shares[0].Amount != 33.34 || shares[1].Amount != 33.33 || shares[2].Amount != 33.33 {
		t.Fatalf("unexpected shares: %#v", shares)
	}
}

func TestEqualSharesSkipsEmptyAndDuplicateParticipants(t *testing.T) {
	shares := equalShares(90, []string{"alex", "", "ivan", "alex"})
	if len(shares) != 2 {
		t.Fatalf("expected 2 unique non-empty shares, got %d", len(shares))
	}
	if got := sharesTotalCents(shares); got != 9000 {
		t.Fatalf("expected total 9000 cents, got %d", got)
	}
}

func TestPaymentAndShareTotalsUseCents(t *testing.T) {
	payments := []models.ExpensePayment{{UserID: "alex", Amount: 10.005}, {UserID: "ivan", Amount: 20.004}}
	shares := []models.ExpenseShare{{UserID: "alex", Amount: 15.005}, {UserID: "ivan", Amount: 15.004}}

	if got := paymentsTotalCents(payments); got != 3001 {
		t.Fatalf("expected payments 3001 cents, got %d", got)
	}
	if got := sharesTotalCents(shares); got != 3001 {
		t.Fatalf("expected shares 3001 cents, got %d", got)
	}
}

func TestDuplicatePaymentsAndShares(t *testing.T) {
	payments := []models.ExpensePayment{{UserID: "alex", Amount: 100}, {UserID: "alex", Amount: 50}}
	shares := []models.ExpenseShare{{UserID: "ivan", Amount: 80}, {UserID: "ivan", Amount: 70}}

	if !hasDuplicatePayments(payments) {
		t.Fatal("expected duplicate payments to be detected")
	}
	if !hasDuplicateShares(shares) {
		t.Fatal("expected duplicate shares to be detected")
	}
}
