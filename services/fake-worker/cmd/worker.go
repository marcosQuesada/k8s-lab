package cmd

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/gorilla/mux"
	cfg "github.com/marcosQuesada/k8s-lab/pkg/config"
	ht "github.com/marcosQuesada/k8s-lab/pkg/http/handler"
	app2 "github.com/marcosQuesada/k8s-lab/services/fake-worker/internal/app"
	cfg2 "github.com/marcosQuesada/k8s-lab/services/fake-worker/internal/config"
	htv "github.com/marcosQuesada/k8s-lab/services/fake-worker/internal/transport/http"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

const DefaultHostName = "swarm-worker-0"

type Processor interface {
	Assign(w *cfg.Workload) error
}

// workerCmd represents the worker command
var workerCmd = &cobra.Command{
	Use:   "worker",
	Short: "worker process assignations",
	Long:  `worker process assigned Jobs, support hot reload starting/stopping required consumers pool`,
	Run: func(cmd *cobra.Command, args []string) {
		name := cfg2.HostName(DefaultHostName)
		log.Infof("worker %s started", name)
		app := app2.NewApp()
		if err := updateWorkloadFromConfig(app); err != nil {
			log.Errorf("unable to watch keys, %v", err)
		}

		viper.OnConfigChange(func(e fsnotify.Event) {
			log.Info("Config file changed")
			if err := updateWorkloadFromConfig(app); err != nil {
				log.Errorf("unable to watch keys, %v", err)
			}
		})

		go app.Run() // @TODO: Right now just a mock

		router := mux.NewRouter()
		ch := ht.NewChecker(cfg.Commit, cfg.Date)
		ch.Routes(router)
		vCh := htv.NewVersionChecker(cfg2.NewVersionAdapter(cfg2.HostName(DefaultHostName)))
		vCh.Routes(router)
		srv := &http.Server{
			Addr:         fmt.Sprintf(":%s", cfg.HttpPort),
			Handler:      router,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
		}

		go func(h *http.Server) {
			log.Infof("starting server on port %s", cfg.HttpPort)
			e := h.ListenAndServe()
			if e != nil && e != http.ErrServerClosed {
				log.Fatalf("Could not Listen and server, error %v", e)
			}
		}(srv)

		sigTerm := make(chan os.Signal, 1)
		signal.Notify(sigTerm, syscall.SIGTERM, syscall.SIGINT)
		<-sigTerm

		if err := srv.Close(); err != nil {
			log.Errorf("unexpected error on http server close %v", err)
		}
		app.Terminate()
	},
}

func init() {
	rootCmd.AddCommand(workerCmd)
}

func updateWorkloadFromConfig(mng Processor) error {
	if err := cfg2.LoadConfig(cfg.ConfigFilePath, cfg.ConfigFile); err != nil {
		log.Fatalf("unable to unMarshall config, error %v", err)
	}

	wl, err := cfg2.HostWorkLoad(cfg2.HostName(DefaultHostName))
	if err != nil {
		log.Errorf("unable to get Job set assignation, error %v", err)

		return nil
	}

	if err := mng.Assign(wl); err != nil {
		return fmt.Errorf("unable to start manager, error %v", err)
	}

	return nil
}
