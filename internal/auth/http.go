package auth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/fred-bonn/srlcode/internal/database"
	"github.com/fred-bonn/srlcode/internal/httpx"
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
			httpx.RespondWithError(w, http.StatusUnauthorized, "missing session token", nil)
			return
		}
		user, err := s.db.GetUserBySessionToken(r.Context(), token)
		if err != nil {
			httpx.RespondWithError(w, http.StatusUnauthorized, "invalid or expired session token", err)
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
		httpx.RespondWithError(w, http.StatusConflict, err.Error(), nil)
	case errors.Is(err, ErrInvalidEmail), errors.Is(err, ErrWeakPassword):
		httpx.RespondWithError(w, http.StatusBadRequest, err.Error(), nil)
	case err != nil:
		httpx.RespondWithError(w, http.StatusInternalServerError, "internal error", err)
	default:
		httpx.RespondWithJSON(w, http.StatusCreated, registerResponse{ID: user.ID, Email: user.Email})
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
		httpx.RespondWithError(w, http.StatusUnauthorized, err.Error(), nil)
	case err != nil:
		httpx.RespondWithError(w, http.StatusInternalServerError, "internal error", err)
	default:
		httpx.RespondWithJSON(w, http.StatusOK, loginResponse{Token: token})
	}
}

func decodeCredentials(w http.ResponseWriter, r *http.Request) (credentials, bool) {
	var creds credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		httpx.RespondWithError(w, http.StatusBadRequest, "invalid request body", nil)
		return creds, false
	}
	return creds, true
}
