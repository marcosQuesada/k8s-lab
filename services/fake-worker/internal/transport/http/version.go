package http

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/marcosQuesada/k8s-lab/pkg/config"
	httpPkg "github.com/marcosQuesada/k8s-lab/pkg/http/handler"
	"net/http"

	log "github.com/sirupsen/logrus"
)

type provider interface {
	Version() int64
	Workload() *config.Workload
}

// VersionChecker handles health checker handler, replying commit version and release date
type VersionChecker struct {
	accessor provider
}

// NewVersionChecker builds health checker handler
func NewVersionChecker(p provider) *VersionChecker {
	return &VersionChecker{
		accessor: p,
	}
}

// versionHandler replies current release hash and date
func (a *VersionChecker) versionHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(httpPkg.ContentType, httpPkg.JSONContentType)
	log.Infof("requested version, got %d jobs %v", a.accessor.Version(), a.accessor.Workload())

	var jobs []config.Job
	if a.accessor.Workload() != nil {
		jobs = a.accessor.Workload().Jobs
	}
	wrk := &config.Workload{
		Jobs: jobs,
	}
	if err := json.NewEncoder(w).Encode(wrk); err != nil {
		log.Errorf("Unexpected error Marshalling version, error %v", err)
	}
}

// Routes defines router endpoints
func (a *VersionChecker) Routes(r *mux.Router) {
	r.HandleFunc(`/internal/version`, a.versionHandler)
}
