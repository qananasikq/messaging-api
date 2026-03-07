package main

import (
	"context"
	"database/sql"
	"log"
	"time"

	"messaging-api/internal/config"

	_ "github.com/jackc/pgx/v5/stdlib" // ВАЖНО: регистрирует драйвер pgx
	"github.com/pressly/goose/v3"
)

func main() {
	cfg := config.MustLoad()

	db, err := sql.Open("pgx", cfg.Postgres.DSN)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatal(err)
	}

	if err := goose.UpContext(ctx, db, "migrations"); err != nil {
		log.Fatal(err)
	}

	log.Println("migrations applied")
}
