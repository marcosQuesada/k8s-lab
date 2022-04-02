package cmd

import (
	"fmt"
	cfg "github.com/marcosQuesada/k8s-lab/pkg/config"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
)

const appID = "config-reloader-controller"

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "root",
	Short: "root controller command",
	Long:  `root controller command`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	cfg.SetCoreFlags(rootCmd, appID)
}

func initConfig() {
	if err := cfg.BuildLogger(appID); err != nil {
		log.Fatalf("unable to build logger, error %v", err)
	}
}
