package cmd

import (
	"fmt"
	"github.com/gorilla/mux"
	config2 "github.com/marcosQuesada/k8s-lab/pkg/config"
	ht "github.com/marcosQuesada/k8s-lab/pkg/http/handler"
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

// internalCmd represents the internal command
var internalCmd = &cobra.Command{
	Use:   "internal",
	Short: "swarm internal controller",
	Long:  `swarm internal controller balance configured keys between swarm peers`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Infof("controller internal listening on namespace %s label %s Version %s release date %s http server on port %s", namespace, watchLabel, config2.Commit, config2.Date, config2.HttpPort)

		cl := operator.BuildInternalClient()
		swarmCl := k8s.BuildSwarmInternalClient()
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
		h := crd.NewHandler(m, nil) //@TODO: BROKEN
		p := operator.NewEventProcessorWithCustomInformer(informer.Informer(), h)
		swarmCtl := operator.NewConsumer(p, q)

		stopCh := make(chan struct{})
		go swarmCtl.Run(stopCh)

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
