package learning

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/google/uuid"

	"github.com/fred-bonn/srlcode/internal/auth"
	"github.com/fred-bonn/srlcode/internal/httpx"
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
		httpx.Error(w, http.StatusInternalServerError, "internal error")
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
		httpx.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	challengeID, err := uuid.Parse(req.ChallengeID)
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid challenge_id")
		return
	}
	attempt, err := s.StartAttempt(r.Context(), userID, challengeID)
	switch {
	case errors.Is(err, ErrNotFound):
		httpx.Error(w, http.StatusNotFound, err.Error())
	case errors.Is(err, ErrNotAssigned):
		httpx.Error(w, http.StatusForbidden, err.Error())
	case err != nil:
		log.Printf("start attempt: %v", err)
		httpx.Error(w, http.StatusInternalServerError, "internal error")
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
		httpx.Error(w, http.StatusBadRequest, "invalid attempt id")
		return
	}
	var req saveCodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	err = s.SaveCode(r.Context(), userID, attemptID, req.Code)
	switch {
	case errors.Is(err, ErrNotFound):
		httpx.Error(w, http.StatusNotFound, err.Error())
	case errors.Is(err, ErrAttemptClosed):
		httpx.Error(w, http.StatusConflict, err.Error())
	case err != nil:
		log.Printf("save code: %v", err)
		httpx.Error(w, http.StatusInternalServerError, "internal error")
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
		httpx.Error(w, http.StatusBadRequest, "invalid attempt id")
		return
	}
	var req completeAttemptRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	err = s.CompleteAttempt(r.Context(), userID, attemptID, req.Correct, req.DurationSeconds)
	switch {
	case errors.Is(err, ErrNotFound):
		httpx.Error(w, http.StatusNotFound, err.Error())
	case errors.Is(err, ErrAttemptClosed):
		httpx.Error(w, http.StatusConflict, err.Error())
	case err != nil:
		log.Printf("complete attempt: %v", err)
		httpx.Error(w, http.StatusInternalServerError, "internal error")
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
		httpx.Error(w, http.StatusConflict, err.Error())
	case err != nil:
		log.Printf("complete lesson: %v", err)
		httpx.Error(w, http.StatusInternalServerError, "internal error")
	default:
		w.WriteHeader(http.StatusNoContent)
	}
}

func authenticated(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, "unauthenticated")
		return uuid.UUID{}, false
	}
	return user.ID, true
}
