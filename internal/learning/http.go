package learning

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/google/uuid"

	"github.com/fred-bonn/srlcode/internal/auth"
)

func (s *Service) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /today", s.handleToday)
	mux.HandleFunc("POST /attempts", s.handleStartAttempt)
	mux.HandleFunc("PUT /attempts/{id}/code", s.handleSaveCode)
	mux.HandleFunc("POST /attempts/{id}/complete", s.handleCompleteAttempt)
	mux.HandleFunc("POST /lessons/complete", s.handleCompleteLesson)
	return mux
}

type startAttemptRequest struct {
	ChallengeID string `json:"challenge_id"`
}

type saveCodeRequest struct {
	Code string `json:"code"`
}

type completeAttemptRequest struct {
	Correct         bool  `json:"correct"`
	DurationSeconds int32 `json:"duration_seconds"`
}

func (s *Service) handleToday(w http.ResponseWriter, r *http.Request) {
	userID, ok := authenticated(w, r)
	if !ok {
		return
	}
	view, err := s.Today(r.Context(), userID)
	if err != nil {
		log.Printf("today: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(view)
}

func (s *Service) handleStartAttempt(w http.ResponseWriter, r *http.Request) {
	userID, ok := authenticated(w, r)
	if !ok {
		return
	}
	var req startAttemptRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	challengeID, err := uuid.Parse(req.ChallengeID)
	if err != nil {
		http.Error(w, "invalid challenge_id", http.StatusBadRequest)
		return
	}
	attempt, err := s.StartAttempt(r.Context(), userID, challengeID)
	switch {
	case errors.Is(err, ErrNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	case errors.Is(err, ErrNotAssigned):
		http.Error(w, err.Error(), http.StatusForbidden)
	case err != nil:
		log.Printf("start attempt: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
	default:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(attempt)
	}
}

func (s *Service) handleSaveCode(w http.ResponseWriter, r *http.Request) {
	userID, ok := authenticated(w, r)
	if !ok {
		return
	}
	attemptID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid attempt id", http.StatusBadRequest)
		return
	}
	var req saveCodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	err = s.SaveCode(r.Context(), userID, attemptID, req.Code)
	switch {
	case errors.Is(err, ErrNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	case errors.Is(err, ErrAttemptClosed):
		http.Error(w, err.Error(), http.StatusConflict)
	case err != nil:
		log.Printf("save code: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
	default:
		w.WriteHeader(http.StatusNoContent)
	}
}

func (s *Service) handleCompleteAttempt(w http.ResponseWriter, r *http.Request) {
	userID, ok := authenticated(w, r)
	if !ok {
		return
	}
	attemptID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid attempt id", http.StatusBadRequest)
		return
	}
	var req completeAttemptRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	err = s.CompleteAttempt(r.Context(), userID, attemptID, req.Correct, req.DurationSeconds)
	switch {
	case errors.Is(err, ErrNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	case errors.Is(err, ErrAttemptClosed):
		http.Error(w, err.Error(), http.StatusConflict)
	case err != nil:
		log.Printf("complete attempt: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
	default:
		w.WriteHeader(http.StatusNoContent)
	}
}

func (s *Service) handleCompleteLesson(w http.ResponseWriter, r *http.Request) {
	userID, ok := authenticated(w, r)
	if !ok {
		return
	}
	err := s.CompleteLesson(r.Context(), userID)
	switch {
	case errors.Is(err, ErrLessonAlreadyDoneToday), errors.Is(err, ErrCurriculumDone):
		http.Error(w, err.Error(), http.StatusConflict)
	case err != nil:
		log.Printf("complete lesson: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
	default:
		w.WriteHeader(http.StatusNoContent)
	}
}

func authenticated(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthenticated", http.StatusUnauthorized)
		return uuid.UUID{}, false
	}
	return user.ID, true
}
