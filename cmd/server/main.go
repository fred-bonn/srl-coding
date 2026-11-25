package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/fred-bonn/srlcode/internal/auth"
	"github.com/fred-bonn/srlcode/internal/database"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is not set. Copy .env.example to .env and fill in your connection string.")
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("opening database: %v", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatalf("connecting to database: %v", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/auth/", auth.New(database.New(db)).Routes())

	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = ":8080"
	}
	log.Printf("listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
