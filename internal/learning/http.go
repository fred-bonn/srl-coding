package learning

import (
	"encoding/json"
	"errors"
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
	TopicID     string `json:"topic_id"`
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
		httpx.RespondWithError(w, http.StatusInternalServerError, "internal error", err)
		return
	}
	httpx.RespondWithJSON(w, http.StatusOK, view)
}

func (s *Service) handleStartAttempt(w http.ResponseWriter, r *http.Request) {
	userID, ok := authenticated(w, r)
	if !ok {
		return
	}
	var req startAttemptRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.RespondWithError(w, http.StatusBadRequest, "invalid request body", nil)
		return
	}
	if (req.ChallengeID == "") == (req.TopicID == "") {
		httpx.RespondWithError(w, http.StatusBadRequest, "exactly one of challenge_id or topic_id is required", nil)
		return
	}
	var attempt AttemptView
	if req.TopicID != "" {
		topicID, err := uuid.Parse(req.TopicID)
		if err != nil {
			httpx.RespondWithError(w, http.StatusBadRequest, "invalid topic_id", nil)
			return
		}
		attempt, err = s.StartTopicAttempt(r.Context(), userID, topicID)
		if err != nil {
			respondStartAttemptError(w, err)
			return
		}
		httpx.RespondWithJSON(w, http.StatusCreated, attempt)
		return
	}
	challengeID, err := uuid.Parse(req.ChallengeID)
	if err != nil {
		httpx.RespondWithError(w, http.StatusBadRequest, "invalid challenge_id", nil)
		return
	}
	attempt, err = s.StartAttempt(r.Context(), userID, challengeID)
	if err != nil {
		respondStartAttemptError(w, err)
		return
	}
	httpx.RespondWithJSON(w, http.StatusCreated, attempt)
}

func respondStartAttemptError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		httpx.RespondWithError(w, http.StatusNotFound, err.Error(), nil)
	case errors.Is(err, ErrNotAssigned):
		httpx.RespondWithError(w, http.StatusForbidden, err.Error(), nil)
	default:
		httpx.RespondWithError(w, http.StatusInternalServerError, "internal error", err)
	}
}

func (s *Service) handleSaveCode(w http.ResponseWriter, r *http.Request) {
	userID, ok := authenticated(w, r)
	if !ok {
		return
	}
	attemptID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		httpx.RespondWithError(w, http.StatusBadRequest, "invalid attempt id", nil)
		return
	}
	var req saveCodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.RespondWithError(w, http.StatusBadRequest, "invalid request body", nil)
		return
	}
	err = s.SaveCode(r.Context(), userID, attemptID, req.Code)
	switch {
	case errors.Is(err, ErrNotFound):
		httpx.RespondWithError(w, http.StatusNotFound, err.Error(), nil)
	case errors.Is(err, ErrAttemptClosed):
		httpx.RespondWithError(w, http.StatusConflict, err.Error(), nil)
	case err != nil:
		httpx.RespondWithError(w, http.StatusInternalServerError, "internal error", err)
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
		httpx.RespondWithError(w, http.StatusBadRequest, "invalid attempt id", nil)
		return
	}
	var req completeAttemptRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.RespondWithError(w, http.StatusBadRequest, "invalid request body", nil)
		return
	}
	err = s.CompleteAttempt(r.Context(), userID, attemptID, req.Correct, req.DurationSeconds)
	switch {
	case errors.Is(err, ErrNotFound):
		httpx.RespondWithError(w, http.StatusNotFound, err.Error(), nil)
	case errors.Is(err, ErrAttemptClosed):
		httpx.RespondWithError(w, http.StatusConflict, err.Error(), nil)
	case err != nil:
		httpx.RespondWithError(w, http.StatusInternalServerError, "internal error", err)
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
		httpx.RespondWithError(w, http.StatusConflict, err.Error(), nil)
	case err != nil:
		httpx.RespondWithError(w, http.StatusInternalServerError, "internal error", err)
	default:
		w.WriteHeader(http.StatusNoContent)
	}
}

func authenticated(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		httpx.RespondWithError(w, http.StatusUnauthorized, "unauthenticated", nil)
		return uuid.UUID{}, false
	}
	return user.ID, true
}
