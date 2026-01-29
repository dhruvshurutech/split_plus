package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dhruvsaxena1998/splitplus/internal/app"
	"github.com/dhruvsaxena1998/splitplus/internal/db"
	"github.com/dhruvsaxena1998/splitplus/internal/db/sqlc"
)

func main() {
	ctx := context.Background()

	pool, err := db.NewPool(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	queries := sqlc.New(pool)

	app := app.New(pool, queries)
	server := &http.Server{
		Addr:         ":8080",
		Handler:      app.Router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Println("server running on :8080")
	log.Fatal(server.ListenAndServe())
}
