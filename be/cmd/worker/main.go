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

	// Initialize dependencies for recurring expenses
	recurringExpenseRepo := repository.NewRecurringExpenseRepository(pool, queries)
	expenseRepo := repository.NewExpenseRepository(pool, queries)
	expenseCategoryRepo := repository.NewExpenseCategoryRepository(pool)
	groupActivityRepo := repository.NewGroupActivityRepository(pool)

	// Initialize dependencies for auth cleanup
	userRepo := repository.NewUserRepository(queries)
	sessionRepo := repository.NewSessionRepository(queries)
	pendingUserRepo := repository.NewPendingUserRepository(pool, queries)

	groupActivityService := service.NewGroupActivityService(groupActivityRepo)
	expenseService := service.NewExpenseService(expenseRepo, expenseCategoryRepo, groupActivityService, userRepo, pendingUserRepo)
	recurringExpenseService := service.NewRecurringExpenseService(recurringExpenseRepo, expenseService)
	userService := service.NewUserService(userRepo)

	// Load JWT config from environment
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable is required")
	}

	jwtService := service.NewJWTService(jwtSecret, 168*3600*1e9, 720*3600*1e9) // 7 days, 30 days
	authService := service.NewAuthService(userService, sessionRepo, jwtService, 168*3600*1e9, 720*3600*1e9)

	// Initialize and start workers
	recurringExpenseGen := job.NewRecurringExpenseGenerator(recurringExpenseService)
	recurringExpenseGen.Start(ctx)

	authCleanup := job.NewAuthCleanup(authService)
	authCleanup.Start(ctx)

	log.Println("Workers started:")
	log.Println("  - Recurring expense generator (daily at 2 AM)")
	log.Println("  - Auth cleanup (hourly)")

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down workers...")
	recurringExpenseGen.Stop()
	authCleanup.Stop()
	log.Println("Workers stopped")
}
