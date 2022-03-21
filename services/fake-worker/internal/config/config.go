package config

import (
	"fmt"
	"github.com/marcosQuesada/k8s-lab/pkg/config"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
	"sync"
)

var (
	// workloadsConfig defines static worker Workload Workloads
	workloadsConfig Config
	workloadsMutex  sync.RWMutex
)

// Config loaded app config
type Config struct {
	Workload map[string]*config.Workload `mapstructure:"workloads"`
	Version  int64                       `mapstructure:"version"`
}

func (cfg *Config) HostWorkLoad(host string) (*config.Workload, error) {
	v, ok := cfg.Workload[host]
	if !ok {
		return nil, fmt.Errorf("unable to find host %s keySet config %v", host, cfg)
	}

	log.Infof("Loaded %d keys on Host %s", len(v.Jobs), host)

	return v, nil
}

// HostWorkLoad gets assignation Workload to the Host
func HostWorkLoad(host string) (*config.Workload, error) {
	workloadsMutex.RLock()
	defer workloadsMutex.RUnlock()

	return workloadsConfig.HostWorkLoad(host)
}

// Version reply version config
func Version() int64 {
	workloadsMutex.RLock()
	defer workloadsMutex.RUnlock()

	return workloadsConfig.Version
}

type VersionAdapter struct {
	hostName string
}

func NewVersionAdapter(host string) *VersionAdapter {
	return &VersionAdapter{
		hostName: host,
	}
}

func (v *VersionAdapter) Version() int64 {
	return Version()
}

func (v *VersionAdapter) Workload() *config.Workload {
	w, err := HostWorkLoad(v.hostName)
	if err != nil {
		return nil
	}
	return w
}

func HostName(defaultHostName string) string {
	p := os.Getenv("HOSTNAME")
	if p == "" {
		return defaultHostName
	}

	return p
}

func LoadConfig(configFilePath, configFile string) error {
	viper.AddConfigPath(configFilePath)
	viper.SetConfigName(configFile)
	viper.SetConfigType("yml")

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("unable to load config file %s from path %s, error %v", configFile, configFilePath, err)
	}

	log.Infof("Using config file: %s", viper.ConfigFileUsed())

	workloadsMutex.Lock()
	defer workloadsMutex.Unlock()
	if err := viper.Unmarshal(&workloadsConfig); err != nil {
		return fmt.Errorf("unable to unMarshall config, error %v", err)
	}

	viper.WatchConfig()

	return nil
}
