package cmd

import (
	"fmt"
	"github.com/gorilla/mux"
	cfg "github.com/marcosQuesada/k8s-lab/pkg/config"
	operator2 "github.com/marcosQuesada/k8s-lab/pkg/operator"
	"github.com/marcosQuesada/k8s-lab/pkg/operator/configmap"
	pod2 "github.com/marcosQuesada/k8s-lab/pkg/operator/pod"
	statefulset2 "github.com/marcosQuesada/k8s-lab/pkg/operator/statefulset"
	app2 "github.com/marcosQuesada/k8s-lab/services/swarm-pool-controller/internal/app"
	"github.com/marcosQuesada/k8s-lab/services/swarm-pool-controller/internal/infra/k8s"
	"github.com/marcosQuesada/k8s-lab/services/swarm-pool-controller/internal/infra/k8s/apis/swarm/v1alpha1"
	crd2 "github.com/marcosQuesada/k8s-lab/services/swarm-pool-controller/internal/infra/k8s/crd"
	ht "github.com/marcosQuesada/k8s-lab/services/swarm-pool-controller/internal/infra/transport/http"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
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
	Short: "swarm external controller, useful on development path",
	Long:  `swarm internal controller balance configured keys between swarm peers, useful on development path`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Infof("controller external listening on namespace %s label %s Version %s release date %s http server on port %s", namespace, watchLabel, cfg.Commit, cfg.Date, cfg.HttpPort)

		cl := operator2.BuildExternalClient()
		swcl := k8s.BuildSwarmExternalClient()
		cm := configmap.NewProvider(cl, namespace, workersConfigMapName, watchLabel)
		vst := ht.NewVersionProvider(cfg.HttpPort) // @TODO: REFACTOR AND REMOVE
		pdl := pod2.NewProvider(cl, namespace)

		swl := crd2.NewProvider(swcl, namespace, watchLabel)
		mex := crd2.NewProviderMiddleware(cm, swl)

		ex := app2.NewExecutor(mex, vst, pdl)
		st := app2.NewState(config.Jobs, watchLabel)
		app := app2.NewWorkerPool(st, ex)

		podLwa := pod2.NewListWatcherAdapter(cl, namespace)
		podH := pod2.NewHandler(app)
		podEventQueue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
		podEventHandler := operator2.NewResourceEventHandler(podEventQueue)
		podEh := operator2.NewLabelSelectorMiddleware(watchLabel, podEventHandler)
		podEvp := operator2.NewEventProcessor(&apiv1.Pod{}, podLwa, podEh, podH)
		podCtl := operator2.NewController(podEvp, podEventQueue)

		stsLwa := statefulset2.NewListWatcherAdapter(cl, namespace)
		stsH := statefulset2.NewHandler(app)
		stsEventQueue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
		stsEventHandler := operator2.NewResourceEventHandler(stsEventQueue)
		stsEh := operator2.NewLabelSelectorMiddleware(watchLabel, stsEventHandler)
		stsEvp := operator2.NewEventProcessor(&appsv1.StatefulSet{}, stsLwa, stsEh, stsH)
		stsCtl := operator2.NewController(stsEvp, stsEventQueue)

		swarmLwa := crd2.NewListWatcherAdapter(swcl, namespace)
		swarmH := crd2.NewHandler()
		swarmEventQueue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
		swarmEventHandler := operator2.NewResourceEventHandler(swarmEventQueue)
		swarmEvp := operator2.NewEventProcessor(&v1alpha1.Swarm{}, swarmLwa, swarmEventHandler, swarmH)
		swarmCtl := operator2.NewController(swarmEvp, swarmEventQueue)

		stopCh := make(chan struct{})
		go podCtl.Run(stopCh)
		go stsCtl.Run(stopCh)
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
		log.Info("Stopping controller")
	},
}

func init() {
	rootCmd.AddCommand(externalCmd)
}
