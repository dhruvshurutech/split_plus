package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/dhruvsaxena1998/splitplus/internal/db"
	"github.com/dhruvsaxena1998/splitplus/internal/db/sqlc"
	"github.com/dhruvsaxena1998/splitplus/internal/job"
	"github.com/dhruvsaxena1998/splitplus/internal/repository"
	"github.com/dhruvsaxena1998/splitplus/internal/service"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup database connection
	pool, err := db.NewPool(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	queries := sqlc.New(pool)

	// Initialize dependencies
	recurringExpenseRepo := repository.NewRecurringExpenseRepository(pool, queries)
	expenseRepo := repository.NewExpenseRepository(pool, queries)
	expenseCategoryRepo := repository.NewExpenseCategoryRepository(pool)
	groupActivityRepo := repository.NewGroupActivityRepository(pool)

	groupActivityService := service.NewGroupActivityService(groupActivityRepo)
	expenseService := service.NewExpenseService(expenseRepo, expenseCategoryRepo, groupActivityService)
	recurringExpenseService := service.NewRecurringExpenseService(recurringExpenseRepo, expenseService)

	// Initialize and start worker
	generator := job.NewRecurringExpenseGenerator(recurringExpenseService)
	generator.Start(ctx)

	log.Println("Recurring expense worker started")

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down worker...")
	generator.Stop()
	log.Println("Worker stopped")
}
