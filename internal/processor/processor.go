package processor

import (
	"context"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yourname/finops-service/internal/models"
	"github.com/yourname/finops-service/internal/repositories"
)

type Processor struct {
	pool        *pgxpool.Pool
	txRepo      repositories.TransactionRepository
	userRepo    repositories.UserRepository
	jobs        chan models.Transaction
	userMutexes map[int]*sync.Mutex
	mu          sync.RWMutex
}

func NewProcessor(pool *pgxpool.Pool, numWorkers int) *Processor {
	p := &Processor{
		pool:        pool,
		txRepo:      repositories.NewTransactionRepo(pool),
		userRepo:    repositories.NewUserRepo(pool),
		jobs:        make(chan models.Transaction, 100),
		userMutexes: make(map[int]*sync.Mutex),
	}

	for i := 0; i < numWorkers; i++ {
		go p.worker()
	}

	return p
}

func (p *Processor) Submit(tx models.Transaction) {
	p.jobs <- tx
}

func (p *Processor) worker() {
	for tx := range p.jobs {
		p.process(tx)
	}
}

func (p *Processor) process(tx models.Transaction) error {
	mu := p.getMutex(tx.UserID)
	mu.Lock()
	defer mu.Unlock()

	dbTx, err := p.pool.Begin(context.Background())
	if err != nil {
		return err
	}
	defer dbTx.Rollback(context.Background())

	balance, err := p.userRepo.GetBalance(context.Background(), tx.UserID)
	if err != nil {
		return err
	}

	if tx.Type == "withdraw" && balance.LessThan(tx.Amount) {
		return nil
	}

	if tx.Type == "deposit" {
		balance = balance.Add(tx.Amount)
	} else {
		balance = balance.Sub(tx.Amount)
	}

	err = p.userRepo.UpdateBalance(context.Background(), tx.UserID, balance, dbTx)
	if err != nil {
		return err
	}

	err = p.txRepo.UpdateTransaction(context.Background(), tx.ID, &models.Transaction{Processed: true})
	if err != nil {
		return err
	}

	return dbTx.Commit(context.Background())
}

func (p *Processor) getMutex(userID int) *sync.Mutex {
	p.mu.RLock()
	mu, ok := p.userMutexes[userID]
	p.mu.RUnlock()

	if ok {
		return mu
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if mu, ok := p.userMutexes[userID]; ok {
		return mu
	}

	mu = &sync.Mutex{}
	p.userMutexes[userID] = mu
	return mu
}

func (p *Processor) Close() error {
	close(p.jobs)
	return nil
}
