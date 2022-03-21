package config

import (
	"testing"
)

func TestItLoadsConfigFromFileAndDumpsSlaveKeySetConfigByHostName(t *testing.T) {
	path := "./"
	configFile := "config.yml"

	if err := LoadConfig(path, configFile); err != nil {
		t.Fatalf("unable to load config file %s from path %s, error %v", configFile, path, err)
	}

	host := HostName("swarm-worker-0")
	if expected, got := "swarm-worker-0", host; expected != got {
		t.Fatalf("host value do not match, expected %s got %s", expected, got)
	}

	keySet, err := workloadsConfig.HostWorkLoad(host)
	if err != nil {
		t.Fatalf("unexpected error getting Workload, error %v", err)
	}

	if expected, got := 12, len(keySet.Jobs); expected != got {
		t.Errorf("expected keys do not match, expected %d got %d", expected, got)
	}
}
