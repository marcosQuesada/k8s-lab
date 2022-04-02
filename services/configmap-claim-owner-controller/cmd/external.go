package cmd

import (
	"fmt"
	"github.com/gorilla/mux"
	cfg "github.com/marcosQuesada/k8s-lab/pkg/config"
	"github.com/marcosQuesada/k8s-lab/pkg/operator"
	"github.com/marcosQuesada/k8s-lab/services/configmap-claim-owner-controller/internal/infra/k8s"
	crdinformers "github.com/marcosQuesada/k8s-lab/services/configmap-claim-owner-controller/internal/infra/k8s/crd/generated/informers/externalversions"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
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
		log.Infof("%s external Version %s release date %s http server on port %s", appID, cfg.Commit, cfg.Date, cfg.HttpPort)

		q := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
		crdClient := k8s.BuildConfigMapClaimOwnerExternalClient()
		informerFactory := crdinformers.NewSharedInformerFactory(crdClient, 0) // @TODO: time.Minute*1)
		eh := operator.NewResourceEventHandler(q)
		informer := informerFactory.K8slab().V1alpha1().ConfigMapClaimOwners()
		informer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
			AddFunc:    eh.Add,
			UpdateFunc: eh.Update,
			DeleteFunc: eh.Delete,
		})

		stopCh := make(chan struct{})
		informerFactory.Start(stopCh)

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
