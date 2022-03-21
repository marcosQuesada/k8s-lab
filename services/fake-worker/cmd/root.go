package cmd

import (
	"fmt"
	cfg "github.com/marcosQuesada/k8s-lab/pkg/config"
	cfg2 "github.com/marcosQuesada/k8s-lab/services/fake-worker/internal/config"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
)

const appID = "swarm-worker"

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "root",
	Short: "root controller command",
	Long:  `root controller command`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if err := cfg.BuildLogger(appID); err != nil {
		log.Fatalf("unable to unMarshall config, error %v", err)
	}

	if err := cfg2.LoadConfig(cfg.ConfigFilePath, cfg.ConfigFile); err != nil {
		log.Fatalf("unable to unMarshall config, error %v", err)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	cfg.SetCoreFlags(rootCmd, "fake-worker")
}
