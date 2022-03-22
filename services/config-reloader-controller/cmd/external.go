package cmd

import (
	"fmt"
	"github.com/gorilla/mux"
	cfg "github.com/marcosQuesada/k8s-lab/pkg/config"
	"github.com/marcosQuesada/k8s-lab/pkg/operator"
	cm "github.com/marcosQuesada/k8s-lab/pkg/operator/configmap"
	"github.com/marcosQuesada/k8s-lab/services/config-reloader-controller/internal/infra/k8s/configmap"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

// externalCmd represents the external command
var externalCmd = &cobra.Command{
	Use:   "external",
	Short: "config reloader external controller, useful on development path",
	Long:  `config reloader controller restarts deployment/statefulset on watched configmap change`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Infof("config reloader controller external listening on namespace %s label %s Version %s release date %s http server on port %s", namespace, watchLabel, cfg.Commit, cfg.Date, cfg.HttpPort)

		cl := operator.BuildExternalClient()

		lwa := cm.NewListWatcherAdapter(cl, namespace)
		h := configmap.NewHandler()
		ctl := operator.Build(lwa, h) // @TODO: watchedConfigMapName

		stopCh := make(chan struct{})
		go ctl.Run(stopCh)

		router := mux.NewRouter()
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
		close(stopCh)
		log.Info("Stopping controller")
	},
}

func init() {
	rootCmd.AddCommand(externalCmd)
}
