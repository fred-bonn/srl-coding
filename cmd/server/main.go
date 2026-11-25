package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/google/uuid"

	"github.com/fred-bonn/srlcode/internal/auth"
	"github.com/fred-bonn/srlcode/internal/challenge"
	"github.com/fred-bonn/srlcode/internal/challengegen"
	"github.com/fred-bonn/srlcode/internal/database"
	"github.com/fred-bonn/srlcode/internal/learning"
	"github.com/fred-bonn/srlcode/internal/lessons"
	"github.com/fred-bonn/srlcode/internal/srs"
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

	queries := database.New(db)

	lessonDir := os.Getenv("LESSONS_DIR")
	if lessonDir == "" {
		lessonDir = "data/lessons/python"
	}
	loaded, err := lessons.Load(lessonDir)
	if err != nil {
		log.Fatalf("loading lessons: %v", err)
	}
	if err := lessons.Sync(context.Background(), queries, loaded); err != nil {
		log.Fatalf("syncing lesson challenges: %v", err)
	}
	log.Printf("synced %d lessons", len(loaded))

	authSvc := auth.New(queries)
	picker := challenge.NewPicker(queries, newGenerator())
	learningSvc := learning.New(queries, srs.Fixed{}, picker, loaded, testUserPredicate(context.Background(), queries))

	mux := http.NewServeMux()
	mux.Handle("/auth/", authSvc.Routes())
	mux.Handle("/", authSvc.Authenticate(learningSvc.Routes()))

	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = ":8080"
	}
	log.Printf("listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

// newGenerator prefers the OpenCode generator, then OpenRouter. With no
// API key set, reviews reuse existing challenges instead of generating.
func newGenerator() challengegen.Generator {
	gen, err := challengegen.NewOpencodeGenerator()
	if err == nil {
		return gen
	}
	if gen, err = challengegen.NewOpenRouterGenerator(); err == nil {
		return gen
	}
	log.Print("no challenge generator configured; reviews will reuse existing challenges")
	return nil
}

// testUserPredicate resolves the TEST_USERS email list at boot; users
// registered later need a restart to be recognized.
func testUserPredicate(ctx context.Context, q database.Querier) func(uuid.UUID) bool {
	testUsers := map[uuid.UUID]struct{}{}
	for _, email := range strings.Split(os.Getenv("TEST_USERS"), ",") {
		email = strings.TrimSpace(email)
		if email == "" {
			continue
		}
		user, err := q.GetUserByEmail(ctx, email)
		if err != nil {
			log.Printf("TEST_USERS: unknown email %q, skipping", email)
			continue
		}
		testUsers[user.ID] = struct{}{}
	}
	if len(testUsers) == 0 {
		return func(uuid.UUID) bool { return false }
	}
	log.Printf("test users enabled for %d account(s)", len(testUsers))
	return func(id uuid.UUID) bool {
		_, ok := testUsers[id]
		return ok
	}
}
