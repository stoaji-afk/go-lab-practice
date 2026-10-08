package main

import (
	"context"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yourname/finops-service/internal/api"
	"github.com/yourname/finops-service/internal/config"
	"github.com/yourname/finops-service/internal/db"
	"github.com/yourname/finops-service/internal/middleware"
	"github.com/yourname/finops-service/internal/processor"
	"github.com/yourname/finops-service/internal/repositories"
	"github.com/yourname/finops-service/internal/services"
)

func main() {
	cfg := config.LoadConfig()

	pool, err := db.NewPool(cfg.DBDSN)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	userRepo := repositories.NewUserRepo(pool)
	txRepo := repositories.NewTransactionRepo(pool)

	userService := services.NewUserService(userRepo)
	txService := services.NewTransactionService(txRepo, userRepo, pool)

	proc := processor.NewProcessor(pool, 5)

	mux := http.NewServeMux()

	mux.HandleFunc("/transactions", api.CreateTransactionHandler(txService, proc))
	mux.HandleFunc("/transactions/", api.GetTransactionHandler(txService))
	mux.HandleFunc("/transactions/", api.UpdateTransactionHandler(txService, proc))
	mux.HandleFunc("/transactions/", api.DeleteTransactionHandler(txService, proc))
	mux.HandleFunc("/users/", api.GetUserBalanceHandler(userService))
	mux.HandleFunc("/health", api.HealthCheckHandler())

	handler := middleware.RecoveryMiddleware(middleware.LoggingMiddleware(mux))

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: handler,
	}

	go func() {
		log.Printf("Server starting on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	go func() {
		log.Printf("pprof starting on :6060")
		if err := http.ListenAndServe(":6060", nil); err != nil {
			log.Printf("pprof server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	proc.Close()
	log.Println("Server stopped")
}
