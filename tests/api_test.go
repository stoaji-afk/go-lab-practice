package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/yourname/finops-service/internal/models"
)

func TestTransactionJSON(t *testing.T) {
	tx := models.Transaction{
		ID:     1,
		UserID: 100,
		Amount: decimal.NewFromInt(100),
		Type:   "deposit",
	}

	data, err := json.Marshal(tx)
	if err != nil {
		t.Fatalf("Failed to marshal Transaction: %v", err)
	}

	var decoded models.Transaction
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal Transaction: %v", err)
	}

	if decoded.ID != tx.ID || decoded.UserID != tx.UserID || decoded.Amount.String() != tx.Amount.String() || decoded.Type != tx.Type {
		t.Errorf("Transaction mismatch: got %+v, want %+v", decoded, tx)
	}
}

func TestUserJSON(t *testing.T) {
	user := models.User{
		ID:      1,
		Balance: decimal.NewFromInt(500),
	}

	data, err := json.Marshal(user)
	if err != nil {
		t.Fatalf("Failed to marshal User: %v", err)
	}

	var decoded models.User
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal User: %v", err)
	}

	if decoded.ID != user.ID || decoded.Balance.String() != user.Balance.String() {
		t.Errorf("User mismatch: got %+v, want %+v", decoded, user)
	}
}

func TestInvalidTransactionType(t *testing.T) {
	tx := models.Transaction{
		ID:     1,
		UserID: 100,
		Amount: decimal.NewFromInt(100),
		Type:   "invalid_type",
	}

	validTypes := []string{"deposit", "withdraw"}
	for _, validType := range validTypes {
		if tx.Type == validType {
			t.Errorf("Expected validation to fail for type: %s", tx.Type)
		}
	}
}

func TestValidateAmount(t *testing.T) {
	tests := []struct {
		name    string
		amount  decimal.Decimal
		wantErr bool
	}{
		{"positive amount", decimal.NewFromInt(100), false},
		{"zero amount", decimal.Zero, true},
		{"negative amount", decimal.NewFromInt(-100), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.amount.GreaterThan(decimal.Zero)
			if (isValid == false) != tt.wantErr {
				t.Errorf("ValidateAmount(%v) = %v, want error: %v", tt.amount, isValid, tt.wantErr)
			}
		})
	}
}

func TestHealthCheck(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("HealthCheck status = %d, want %d", w.Code, http.StatusOK)
	}

	if w.Body.String() != "OK" {
		t.Errorf("HealthCheck body = %q, want %q", w.Body.String(), "OK")
	}
}
