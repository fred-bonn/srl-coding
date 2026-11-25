package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"

	"github.com/fred-bonn/srlcode/internal/database"
)

const sessionTTL = 30 * 24 * time.Hour

var (
	ErrEmailTaken         = errors.New("email already registered")
	ErrInvalidEmail       = errors.New("invalid email")
	ErrWeakPassword       = errors.New("password must be at least 8 characters")
	ErrInvalidCredentials = errors.New("invalid email or password")
)

type Service struct {
	db  database.Querier
	now func() time.Time
}

func New(db database.Querier) *Service {
	return &Service{db: db, now: time.Now}
}

func (s *Service) Register(ctx context.Context, email, password string) (database.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if !strings.Contains(email, "@") {
		return database.User{}, ErrInvalidEmail
	}
	if len(password) < 8 {
		return database.User{}, ErrWeakPassword
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return database.User{}, fmt.Errorf("hashing password: %w", err)
	}
	user, err := s.db.CreateUser(ctx, database.CreateUserParams{Email: email, PasswordHash: string(hash)})
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return database.User{}, ErrEmailTaken
	}
	if err != nil {
		return database.User{}, fmt.Errorf("creating user: %w", err)
	}
	return user, nil
}

func (s *Service) Login(ctx context.Context, email, password string) (string, error) {
	user, err := s.db.GetUserByEmail(ctx, strings.ToLower(strings.TrimSpace(email)))
	if err != nil {
		return "", ErrInvalidCredentials
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		return "", ErrInvalidCredentials
	}
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("generating token: %w", err)
	}
	token := hex.EncodeToString(raw)
	if err := s.db.CreateAuthSession(ctx, database.CreateAuthSessionParams{
		Token:     token,
		UserID:    user.ID,
		ExpiresAt: s.now().Add(sessionTTL),
	}); err != nil {
		return "", fmt.Errorf("creating session: %w", err)
	}
	return token, nil
}
