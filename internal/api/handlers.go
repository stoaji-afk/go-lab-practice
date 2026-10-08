package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/shopspring/decimal"
	"github.com/yourname/finops-service/internal/models"
	"github.com/yourname/finops-service/internal/processor"
	"github.com/yourname/finops-service/internal/services"
)

type TransactionRequest struct {
	UserID int             `json:"user_id"`
	Amount decimal.Decimal `json:"amount"`
	Type   string          `json:"type"`
}

func CreateTransactionHandler(txService *services.TransactionService, p *processor.Processor) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req TransactionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		tx := &models.Transaction{
			UserID: req.UserID,
			Amount: req.Amount,
			Type:   req.Type,
		}

		id, err := txService.CreateTransaction(r.Context(), tx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		p.Submit(*tx)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]int{"id": id})
	}
}

func GetTransactionHandler(txService *services.TransactionService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := strings.TrimPrefix(r.URL.Path, "/transactions/")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}

		tx, err := txService.GetTransaction(r.Context(), id)
		if err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(tx)
	}
}

func UpdateTransactionHandler(txService *services.TransactionService, p *processor.Processor) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := strings.TrimPrefix(r.URL.Path, "/transactions/")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}

		var req TransactionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		newTx := &models.Transaction{
			Type:   req.Type,
			Amount: req.Amount,
		}

		if err := txService.UpdateTransaction(r.Context(), id, newTx); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		p.Submit(*newTx)

		w.WriteHeader(http.StatusNoContent)
	}
}

func DeleteTransactionHandler(txService *services.TransactionService, p *processor.Processor) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := strings.TrimPrefix(r.URL.Path, "/transactions/")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}

		if err := txService.DeleteTransaction(r.Context(), id); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		p.Submit(models.Transaction{ID: id})

		w.WriteHeader(http.StatusNoContent)
	}
}

func GetUserBalanceHandler(userService *services.UserService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(r.URL.Path, "/")
		userIDStr := parts[len(parts)-2]
		userID, err := strconv.Atoi(userIDStr)
		if err != nil {
			http.Error(w, "invalid user_id", http.StatusBadRequest)
			return
		}

		balance, err := userService.GetBalance(r.Context(), userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"balance": balance.String()})
	}
}

func HealthCheckHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}
}
