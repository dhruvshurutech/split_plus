package job

import (
	"context"
	"log"
	"time"

	"github.com/dhruvsaxena1998/splitplus/internal/service"
)

type RecurringExpenseGenerator struct {
	service service.RecurringExpenseService
	ticker  *time.Ticker
	done    chan bool
}

func NewRecurringExpenseGenerator(service service.RecurringExpenseService) *RecurringExpenseGenerator {
	return &RecurringExpenseGenerator{
		service: service,
		done:    make(chan bool),
	}
}

func (g *RecurringExpenseGenerator) Start(ctx context.Context) {
	// Run daily at 2 AM
	// Calculate duration until next 2 AM
	now := time.Now()
	nextRun := time.Date(now.Year(), now.Month(), now.Day(), 2, 0, 0, 0, now.Location())
	if nextRun.Before(now) {
		nextRun = nextRun.Add(24 * time.Hour)
	}
	initialDelay := nextRun.Sub(now)

	// Start ticker for daily runs
	g.ticker = time.NewTicker(24 * time.Hour)

	// Wait for initial delay, then start processing
	go func() {
		select {
		case <-time.After(initialDelay):
			// First run
			g.processDueExpenses(ctx)
		case <-ctx.Done():
			return
		case <-g.done:
			return
		}
	}()

	// Process on ticker
	go func() {
		for {
			select {
			case <-g.ticker.C:
				g.processDueExpenses(ctx)
			case <-ctx.Done():
				return
			case <-g.done:
				return
			}
		}
	}()
}

func (g *RecurringExpenseGenerator) Stop() {
	if g.ticker != nil {
		g.ticker.Stop()
	}
	close(g.done)
}

func (g *RecurringExpenseGenerator) processDueExpenses(ctx context.Context) {
	log.Println("Processing due recurring expenses...")

	err := g.service.ProcessDueRecurringExpenses(ctx)
	if err != nil {
		log.Printf("Error processing due recurring expenses: %v", err)
		return
	}

	log.Println("Finished processing due recurring expenses")
}
