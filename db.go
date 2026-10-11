package main

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// checkDatabase temporarily checks connectivity; it does not store jobs yet.
func openDatabase() (*sql.DB, error) {
	db, err := sql.Open("pgx", "postgres://jobrunner:localdev@127.0.0.1:5432/jobrunner?sslmode=disable")
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}
