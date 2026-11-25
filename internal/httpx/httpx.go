// Package httpx holds shared HTTP response helpers.
package httpx

import (
	"encoding/json"
	"log"
	"net/http"
)

func RespondWithJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// RespondWithError sends message to the client as a JSON error; err is
// only logged and may be nil.
func RespondWithError(w http.ResponseWriter, code int, message string, err error) {
	if err != nil {
		log.Printf("%d %s: %s", code, message, err)
	}
	RespondWithJSON(w, code, struct {
		Error string `json:"error"`
	}{message})
}
