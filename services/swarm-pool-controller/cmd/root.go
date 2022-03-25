package cmd

import (
	"fmt"
	cfg "github.com/marcosQuesada/k8s-lab/pkg/config"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

const appID = "swarm-pool-controller"

var (
	config               *cfg.Workload
	namespace            string
	watchLabel           string
	workersConfigMapName string
)

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
		log.Fatalf("unable to build logger, error %v", err)
	}

	viper.AddConfigPath(cfg.ConfigFilePath)
	viper.SetConfigName(cfg.ConfigFile)
	viper.SetConfigType("yml")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("unable to read config, error %v", err)
	}
	if err := viper.Unmarshal(&config); err != nil {
		log.Fatalf("unable to unMarshall config, error %v", err)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	cfg.SetCoreFlags(rootCmd, "swarm-pool-controller")

	rootCmd.PersistentFlags().StringVar(&namespace, "namespace", "swarm", "namespace to listen")
	if p := os.Getenv("NAMESPACE"); p != "" {
		namespace = p
	}

	rootCmd.PersistentFlags().StringVar(&watchLabel, "label", "swarm-worker", "label to watch statefulsets and pods")
	if p := os.Getenv("WATCHED_LABEL"); p != "" {
		watchLabel = p
	}

	rootCmd.PersistentFlags().StringVar(&workersConfigMapName, "configmap", "swarm-worker-config", "workers configmap name")
	if p := os.Getenv("WORKERS_CONFIGMAP_NAME"); p != "" {
		workersConfigMapName = p
	}
}
