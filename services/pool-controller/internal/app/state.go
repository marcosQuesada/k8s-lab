package app

import (
	"fmt"
	"github.com/marcosQuesada/k8s-lab/pkg/config"
	log "github.com/sirupsen/logrus"
	"sync"
)

type state struct {
	setName string
	jobs    []config.Job
	config  *config.Workloads
	mutex   sync.RWMutex
}

// NewState holds workload assignations in the workers pool
func NewState(keySet []config.Job, setName string) *state {
	return &state{
		jobs:    keySet,
		setName: setName,
		config:  &config.Workloads{Workloads: map[string]*config.Workload{}},
	}
}

// BalanceWorkload balances configured workload between workers
func (s *state) BalanceWorkload(totalWorkers int, version int64) error {
	log.Infof("State balance started, Recalculate assignations total workers: %d", totalWorkers)

	s.mutex.Lock()
	defer s.mutex.Unlock()

	if totalWorkers == 0 {
		s.cleanAssignations(0)
		return nil
	}

	partSize := len(s.jobs) / totalWorkers
	modulePartSize := len(s.jobs) % totalWorkers
	for i := 0; i < totalWorkers; i++ {
		workerName := fmt.Sprintf("%s-%d", s.setName, i)
		if _, ok := s.config.Workloads[workerName]; !ok {
			s.config.Workloads[workerName] = &config.Workload{}
		}

		size := partSize
		if i < modulePartSize {
			size++
		}
		start := i * size
		end := (i + 1) * size
		if end > len(s.jobs) {
			end = len(s.jobs)
		}

		s.config.Workloads[workerName] = &config.Workload{Jobs: s.jobs[start:end]}

		log.Infof("worker %s total jobs %d", workerName, len(s.jobs[start:end]))
	}

	s.config.Version = version

	// on Downscaling
	if totalWorkers < len(s.config.Workloads) {
		s.cleanAssignations(totalWorkers)
	}

	return nil
}

// Workloads returns last computed workloads assignations
func (s *state) Workloads() *config.Workloads {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.config
}

// Workload returns concrete workload
func (s *state) Workload(workerIdx int) (*config.Workload, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	asg, ok := s.config.Workloads[fmt.Sprintf("%s-%d", s.setName, workerIdx)]
	if !ok {
		return nil, fmt.Errorf("Workloads not found on index %d", workerIdx)
	}

	return asg, nil
}

func (s *state) cleanAssignations(totalWorkers int) {
	orgSize := len(s.config.Workloads)
	for i := totalWorkers; i < orgSize; i++ {
		delete(s.config.Workloads, fmt.Sprintf("%s-%d", s.setName, i))
	}
}

func (s *state) size() int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return len(s.config.Workloads)
}
