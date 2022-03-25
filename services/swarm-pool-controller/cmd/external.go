package cmd

import (
	"fmt"
	"github.com/gorilla/mux"
	cfg "github.com/marcosQuesada/k8s-lab/pkg/config"
	"github.com/marcosQuesada/k8s-lab/pkg/operator"
	"github.com/marcosQuesada/k8s-lab/pkg/operator/configmap"
	"github.com/marcosQuesada/k8s-lab/pkg/operator/pod"
	"github.com/marcosQuesada/k8s-lab/services/swarm-pool-controller/internal/app"
	"github.com/marcosQuesada/k8s-lab/services/swarm-pool-controller/internal/infra/k8s"
	"github.com/marcosQuesada/k8s-lab/services/swarm-pool-controller/internal/infra/k8s/crd"
	crdinformers "github.com/marcosQuesada/k8s-lab/services/swarm-pool-controller/internal/infra/k8s/crd/generated/informers/externalversions"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// externalCmd represents the external command
var externalCmd = &cobra.Command{
	Use:   "external",
	Short: "swarm pool external controller, useful on development path",
	Long:  `swarm pool internal controller balance configured keys between swarm peers, useful on development path`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Infof("controller external listening on namespace %s label %s Version %s release date %s http server on port %s", namespace, watchLabel, cfg.Commit, cfg.Date, cfg.HttpPort)

		cl := operator.BuildExternalClient()
		swarmCl := k8s.BuildSwarmExternalClient()
		cm := configmap.NewProvider(cl, namespace, workersConfigMapName, watchLabel)
		podp := pod.NewProvider(cl)
		swl := crd.NewProvider(swarmCl, namespace, watchLabel)
		mex := crd.NewProviderMiddleware(cm, swl)
		ex := app.NewExecutor(mex, podp)

		q := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
		informerFactory := crdinformers.NewSharedInformerFactory(swarmCl, 0)
		eh := operator.NewResourceEventHandler(q)
		informer := informerFactory.K8slab().V1alpha1().Swarms()
		informer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
			AddFunc:    eh.Add,
			UpdateFunc: eh.Update,
			DeleteFunc: eh.Delete,
		})

		wpf := k8s.NewWorkerPoolFactory(cl, ex)
		m := app.NewManager(wpf)
		h := crd.NewHandler(m)
		p := operator.NewEventProcessorWithCustomInformer(informer.Informer(), h)
		swarmCtl := operator.NewConsumer(p, q)

		stopCh := make(chan struct{})
		go swarmCtl.Run(stopCh)

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
		_ = srv.Close()
		wpf.Shutdown()
		q.ShutDown()
		log.Info("Stopping controller")
	},
}

func init() {
	rootCmd.AddCommand(externalCmd)
}
