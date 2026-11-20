package main

import (
	"context"
	"database/sql"
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/fred-bonn/srlcode/internal/challengegen"
	"github.com/fred-bonn/srlcode/internal/database"
	_ "github.com/jackc/pgx/v5/stdlib"
)

//go:embed data/*.json
var seedFS embed.FS

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
	if err := seed(db); err != nil {
		log.Fatal(err)
	}
	log.Print("database seeded")
}

func seed(db *sql.DB) error {
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, "TRUNCATE TABLE challenges, topics, challenge_topics RESTART IDENTITY CASCADE"); err != nil {
		return fmt.Errorf("truncating tables: %w", err)
	}

	entries, err := seedFS.ReadDir("data")
	if err != nil {
		return fmt.Errorf("reading seed data: %w", err)
	}
	q := database.New(db).WithTx(tx)
	for _, entry := range entries {
		content, err := seedFS.ReadFile("data/" + entry.Name())
		if err != nil {
			return fmt.Errorf("reading seed file %s: %w", entry.Name(), err)
		}
		var challenge challengegen.Challenge
		if err := json.Unmarshal(content, &challenge); err != nil {
			return fmt.Errorf("parsing seed file %s: %w", entry.Name(), err)
		}
		inserted, err := q.CreateChallenge(ctx, database.CreateChallengeParams{
			Description:  challenge.Description,
			StartCode:    challenge.StartCode,
			TestCode:     challenge.TestCode,
			SolutionCode: challenge.SolutionCode,
			Difficulty:   int32(challenge.Difficulty),
		})
		if err != nil {
			return fmt.Errorf("inserting seed file %s: %w", entry.Name(), err)
		}
		for _, name := range challenge.Topics {
			topic, err := q.CreateTopic(ctx, name)
			if err != nil {
				return fmt.Errorf("inserting topic %s: %w", name, err)
			}
			if err := q.CreateChallengeTopic(ctx, database.CreateChallengeTopicParams{
				ChallengeID: inserted.ID,
				TopicID:     topic.ID,
			}); err != nil {
				return fmt.Errorf("linking challenge %s to topic %s: %w", entry.Name(), name, err)
			}
		}
	}
	return tx.Commit()
}
