package app

import (
	cfg "github.com/marcosQuesada/k8s-lab/pkg/config"
	log "github.com/sirupsen/logrus"
	"sync"
)

type App struct {
	state *cfg.Workload
	mutex sync.Mutex
}

func NewApp() *App {
	return &App{}
}

func (a *App) Assign(w *cfg.Workload) error {
	log.Infof("App Workloads updated Workloads %d", len(w.Jobs))
	if a.state == nil {
		a.state = w
		return nil
	}
	
	i, e := a.state.Difference(w)
	if len(i) == 0 && len(e) == 0 {
		return nil
	}

	log.Infof("Workload State Updated includes %v excludes %v", i, e)
	return nil
}

func (a *App) Run() {
	// @TODO: Pending to implement
}

func (a *App) Terminate() {

}
