package config

import (
	"fmt"
	logger "github.com/marcosQuesada/k8s-lab/pkg/log"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
)

var (
	// Commit hash on current version
	Commit string

	// Date on current release build
	Date string

	LogLevel       string
	Env            string
	ConfigFile     string
	ConfigFilePath string
	HttpPort       string
)

func BuildLogger(appID string) error {
	level, err := log.ParseLevel(LogLevel)
	if err != nil {
		return fmt.Errorf("unexpected error parsing level, error %v", err)
	}
	log.SetLevel(level)
	log.SetReportCaller(true)
	log.SetFormatter(logger.PrettifiedFormatter())
	log.AddHook(logger.NewGlobalFieldHook(appID, Env))

	return nil
}

func SetCoreFlags(cmd *cobra.Command, service string) {
	cmd.PersistentFlags().StringVar(&LogLevel, "log-level", "info", "logging level")
	cmd.PersistentFlags().StringVar(&Env, "env", "dev", "environment where the application is running")
	cmd.PersistentFlags().StringVar(&ConfigFile, "config", "config.yml", "config file source")
	cmd.PersistentFlags().StringVar(&ConfigFilePath, "config-path", fmt.Sprintf("services/%s/config", service), "config file path")
	if p := os.Getenv("CONFIG_PATH"); p != "" {
		ConfigFilePath = p
	}
	cmd.PersistentFlags().StringVar(&HttpPort, "http-port", "9090", "http server port")
	if p := os.Getenv("HTTP_PORT"); p != "" {
		HttpPort = p
	}
}

// Job defines task assignation
type Job string

// Workload definitions from config
type Workload struct {
	Jobs []Job `mapstructure:"jobs" json:"jobs"`
}

// Workloads defines all workload assignations to workers
type Workloads struct {
	Workloads map[string]*Workload `mapstructure:"workloads" json:"workloads"`
	Version   int64                `mapstructure:"version" json:"version"`
}

func (a *Workloads) Equals(asg *Workloads) bool {
	if asg == nil {
		return false
	}

	if asg.Version != a.Version {
		log.Infof("Assignations with different versions, current %d got %d", a.Version, asg.Version)
		return false
	}

	for s, w := range a.Workloads {
		v, ok := asg.Workloads[s]
		if !ok {
			return false
		}
		if !v.Equals(w) {
			return false
		}
	}

	return true
}

func (a *Workload) Equals(wrk *Workload) bool {
	if wrk == nil {
		return false
	}

	if len(a.Jobs) != len(wrk.Jobs) {
		return false
	}

	i, e := a.Difference(wrk)

	return len(i) == 0 && len(e) == 0
}

func (a *Workload) Difference(newWorkload *Workload) (included, excluded []Job) {
	newAssigned := toMap(newWorkload.Jobs)
	for _, k := range a.Jobs {
		if _, ok := newAssigned[string(k)]; !ok {
			excluded = append(excluded, k)
		}
	}

	assigned := toMap(a.Jobs)
	for _, k := range newWorkload.Jobs {
		if _, ok := assigned[string(k)]; !ok {
			included = append(included, k)
		}
	}

	return
}

func toMap(set []Job) map[string]struct{} {
	res := map[string]struct{}{}
	for _, s := range set {
		res[string(s)] = struct{}{}
	}
	return res
}

func jobsEquals(o, t []Job) bool {
	if len(o) != len(t) {
		return false
	}
	theirs := toMap(t)
	for _, k := range o {
		if _, ok := theirs[string(k)]; !ok {
			return false
		}
	}

	return true
}
