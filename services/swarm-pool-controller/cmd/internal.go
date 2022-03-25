package cmd

import (
	"fmt"
	"github.com/gorilla/mux"
	config2 "github.com/marcosQuesada/k8s-lab/pkg/config"
	ht "github.com/marcosQuesada/k8s-lab/pkg/http/handler"
	"github.com/marcosQuesada/k8s-lab/pkg/operator"
	"github.com/marcosQuesada/k8s-lab/pkg/operator/configmap"
	pod2 "github.com/marcosQuesada/k8s-lab/pkg/operator/pod"
	"github.com/marcosQuesada/k8s-lab/pkg/operator/statefulset"
	app2 "github.com/marcosQuesada/k8s-lab/services/swarm-pool-controller/internal/app"
	"github.com/marcosQuesada/k8s-lab/services/swarm-pool-controller/internal/infra/k8s"
	"github.com/marcosQuesada/k8s-lab/services/swarm-pool-controller/internal/infra/k8s/crd"
	"github.com/marcosQuesada/k8s-lab/services/swarm-pool-controller/internal/infra/k8s/pod"
	statefulset2 "github.com/marcosQuesada/k8s-lab/services/swarm-pool-controller/internal/infra/k8s/statefulset"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
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
		pdl := pod2.NewProvider(cl, namespace)
		swl := crd.NewProvider(swarmCl, namespace, watchLabel)
		mex := crd.NewProviderMiddleware(cm, swl)

		ex := app2.NewExecutor(mex, pdl)
		st := app2.NewState(config.Jobs, watchLabel)
		app := app2.NewWorkerPool(st, ex)

		podLwa := pod2.NewListWatcherAdapter(cl, namespace)
		podH := pod.NewHandler(app)
		podCtl := operator.Build(podLwa, podH)

		stsLwa := statefulset.NewListWatcherAdapter(cl, namespace)
		stsH := statefulset2.NewHandler(app)
		stsCtl := operator.Build(stsLwa, stsH)

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
