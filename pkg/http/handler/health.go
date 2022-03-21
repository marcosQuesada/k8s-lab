package handler

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"net/http"

	log "github.com/sirupsen/logrus"
)

// ContentType header
const ContentType = "Content-Type"

// JSONContentType is the json content type
const JSONContentType = "application/json"

// Checker handles health checker handler, replying commit version and release date
type Checker struct {
	version string
	date    string
}

// NewChecker builds health checker handler
func NewChecker(commitVersion, date string) *Checker {
	return &Checker{
		version: commitVersion,
		date:    date,
	}
}

// healthHandler replies current release hash and date
func (a *Checker) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(ContentType, JSONContentType)
	res := map[string]string{"version": a.version, "date": a.date}
	if err := json.NewEncoder(w).Encode(res); err != nil {
		log.Errorf("Unexpected error Marshalling version, error %v", err)
	}
}

// Routes defines router endpoints
func (a *Checker) Routes(r *mux.Router) {
	r.HandleFunc(`/internal/health`, a.healthHandler)
}
