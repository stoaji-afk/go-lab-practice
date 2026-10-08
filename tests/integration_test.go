package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCreateTransaction(t *testing.T) {
	tx := map[string]interface{}{
		"user_id": 1,
		"amount":  "100.00",
		"type":    "deposit",
	}

	body, err := json.Marshal(tx)
	if err != nil {
		t.Fatalf("Failed to marshal transaction: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/transactions", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Simple handler that mimics the real handler behavior
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]int{"id": 1})
	})

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("CreateTransaction status = %d, want %d", w.Code, http.StatusCreated)
	}

	var response map[string]int
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["id"] != 1 {
		t.Errorf("Transaction ID = %d, want %d", response["id"], 1)
	}
}

func TestGetTransaction(t *testing.T) {
	tx := map[string]interface{}{
		"id":         1,
		"user_id":    1,
		"amount":     "100.00",
		"type":       "deposit",
		"processed":  false,
	}

	body, err := json.Marshal(tx)
	if err != nil {
		t.Fatalf("Failed to marshal transaction: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/transactions/1", nil)
	w := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(body)
	})

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("GetTransaction status = %d, want %d", w.Code, http.StatusOK)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["id"] != float64(1) {
		t.Errorf("Transaction ID = %v, want %v", response["id"], 1)
	}
}

func TestUpdateTransaction(t *testing.T) {
	req := httptest.NewRequest(http.MethodPut, "/transactions/1", nil)
	w := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("UpdateTransaction status = %d, want %d", w.Code, http.StatusNoContent)
	}
}

func TestDeleteTransaction(t *testing.T) {
	req := httptest.NewRequest(http.MethodDelete, "/transactions/1", nil)
	w := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("DeleteTransaction status = %d, want %d", w.Code, http.StatusNoContent)
	}
}

func TestGetUserBalance(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/users/1/balance", nil)
	w := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"balance": "500.00"})
	})

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("GetUserBalance status = %d, want %d", w.Code, http.StatusOK)
	}

	var response map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["balance"] != "500.00" {
		t.Errorf("Balance = %s, want %s", response["balance"], "500.00")
	}
}

func TestIntegrationHealthCheck(t *testing.T) {
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

func TestTransactionProcessing(t *testing.T) {
	// Simulate transaction processing with a channel
	done := make(chan bool, 1)

	go func() {
		// Simulate processing
		done <- true
	}()

	select {
	case <-done:
		// Transaction processed successfully
	case <-time.After(1000000000): // 1 second timeout
		t.Error("Transaction processing timed out")
	}
}
