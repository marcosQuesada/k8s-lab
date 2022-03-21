package cmd

import (
	"fmt"
	"github.com/gorilla/mux"
	config2 "github.com/marcosQuesada/k8s-lab/pkg/config"
	ht "github.com/marcosQuesada/k8s-lab/pkg/http/handler"
	operator2 "github.com/marcosQuesada/k8s-lab/pkg/operator"
	"github.com/marcosQuesada/k8s-lab/pkg/operator/configmap"
	"github.com/marcosQuesada/k8s-lab/pkg/operator/crd"
	pod2 "github.com/marcosQuesada/k8s-lab/pkg/operator/pod"
	statefulset2 "github.com/marcosQuesada/k8s-lab/pkg/operator/statefulset"
	app2 "github.com/marcosQuesada/k8s-lab/services/pool-config-controller/internal/app"
	"github.com/marcosQuesada/k8s-lab/services/pool-config-controller/internal/infra/k8s"
	cht "github.com/marcosQuesada/k8s-lab/services/pool-config-controller/internal/infra/transport/http"
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

// internalCmd represents the internal command
var internalCmd = &cobra.Command{
	Use:   "internal",
	Short: "swarm internal controller",
	Long:  `swarm internal controller balance configured keys between swarm peers`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Infof("controller internal listening on namespace %s label %s Version %s release date %s http server on port %s", namespace, watchLabel, config2.Commit, config2.Date, config2.HttpPort)

		cl := operator2.BuildInternalClient()
		swcl := k8s.BuildSwarmInternalClient()
		cm := configmap.NewProvider(cl, namespace, workersConfigMapName, watchLabel)
		vst := cht.NewVersionProvider(config2.HttpPort)
		pdl := pod2.NewProvider(cl, namespace)
		swl := crd.NewProvider(swcl, namespace, watchLabel)
		mex := crd.NewProviderMiddleware(cm, swl)

		ex := app2.NewExecutor(mex, vst, pdl)
		st := app2.NewState(config.Jobs, watchLabel)
		app := app2.NewWorkerPool(st, ex)

		podLwa := pod2.NewListWatcherAdapter(cl, namespace)
		podH := pod2.NewHandler(app)
		podSelector := operator2.NewSelector(watchLabel)
		podEventQueue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
		podEventHandler := operator2.NewResourceEventHandler(podSelector, podEventQueue)
		podEvp := operator2.NewEventProcessor(&apiv1.Pod{}, podLwa, podEventHandler, podH)
		podCtl := operator2.NewController(podEvp, podEventQueue)

		stsLwa := statefulset2.NewListWatcherAdapter(cl, namespace)
		stsH := statefulset2.NewHandler(app)
		stsSelector := operator2.NewSelector(watchLabel)
		stsEventQueue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
		stsEventHandler := operator2.NewResourceEventHandler(stsSelector, stsEventQueue)
		stsEvp := operator2.NewEventProcessor(&appsv1.StatefulSet{}, stsLwa, stsEventHandler, stsH)
		stsCtl := operator2.NewController(stsEvp, stsEventQueue)

		stopCh := make(chan struct{})
		go podCtl.Run(stopCh)
		go stsCtl.Run(stopCh)

		router := mux.NewRouter()
		ch := ht.NewChecker(config2.Commit, config2.Date)
		ch.Routes(router)

		srv := &http.Server{
			Addr:         fmt.Sprintf(":%s", config2.HttpPort),
			Handler:      router,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
		}

		go func(h *http.Server) {
			log.Infof("starting server on port %s", config2.HttpPort)
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
	},
}

func init() {
	rootCmd.AddCommand(internalCmd)
}
