package auth

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/fred-bonn/srlcode/internal/database"
)

type credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type registerResponse struct {
	ID    uuid.UUID `json:"id"`
	Email string    `json:"email"`
}

type loginResponse struct {
	Token string `json:"token"`
}

func (s *Service) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /auth/register", s.handleRegister)
	mux.HandleFunc("POST /auth/login", s.handleLogin)
	return mux
}

type contextKey struct{}

// UserFromContext returns the authenticated user, if any.
func UserFromContext(ctx context.Context) (database.User, bool) {
	user, ok := ctx.Value(contextKey{}).(database.User)
	return user, ok
}

// Authenticate resolves the Bearer session token to a user and passes the
// user to downstream handlers via the request context.
func (s *Service) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := strings.TrimSpace(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
		if token == "" {
			http.Error(w, "missing session token", http.StatusUnauthorized)
			return
		}
		user, err := s.db.GetUserBySessionToken(r.Context(), token)
		if err != nil {
			http.Error(w, "invalid or expired session token", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), contextKey{}, user)))
	})
}

func (s *Service) handleRegister(w http.ResponseWriter, r *http.Request) {
	creds, ok := decodeCredentials(w, r)
	if !ok {
		return
	}
	user, err := s.Register(r.Context(), creds.Email, creds.Password)
	switch {
	case errors.Is(err, ErrEmailTaken):
		http.Error(w, err.Error(), http.StatusConflict)
	case errors.Is(err, ErrInvalidEmail), errors.Is(err, ErrWeakPassword):
		http.Error(w, err.Error(), http.StatusBadRequest)
	case err != nil:
		log.Printf("register: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
	default:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(registerResponse{ID: user.ID, Email: user.Email})
	}
}

func (s *Service) handleLogin(w http.ResponseWriter, r *http.Request) {
	creds, ok := decodeCredentials(w, r)
	if !ok {
		return
	}
	token, err := s.Login(r.Context(), creds.Email, creds.Password)
	switch {
	case errors.Is(err, ErrInvalidCredentials):
		http.Error(w, err.Error(), http.StatusUnauthorized)
	case err != nil:
		log.Printf("login: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
	default:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(loginResponse{Token: token})
	}
}

func decodeCredentials(w http.ResponseWriter, r *http.Request) (credentials, bool) {
	var creds credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return creds, false
	}
	return creds, true
}
