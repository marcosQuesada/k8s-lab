package cmd

import (
	"context"
	"fmt"
	"github.com/gorilla/mux"
	cfg "github.com/marcosQuesada/k8s-lab/pkg/config"
	"github.com/marcosQuesada/k8s-lab/pkg/operator"
	"github.com/marcosQuesada/k8s-lab/pkg/operator/configmap"
	crdop "github.com/marcosQuesada/k8s-lab/pkg/operator/crd"
	"github.com/marcosQuesada/k8s-lab/services/swarm-pool-controller/internal/app"
	"github.com/marcosQuesada/k8s-lab/services/swarm-pool-controller/internal/infra/k8s"
	"github.com/marcosQuesada/k8s-lab/services/swarm-pool-controller/internal/infra/k8s/crd"
	"github.com/marcosQuesada/k8s-lab/services/swarm-pool-controller/internal/infra/k8s/crd/apis/swarm/v1alpha1"
	crdinformers "github.com/marcosQuesada/k8s-lab/services/swarm-pool-controller/internal/infra/k8s/crd/generated/informers/externalversions"
	statefulset "github.com/marcosQuesada/k8s-lab/services/swarm-pool-controller/internal/infra/k8s/statefulset"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
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
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		clientSet := operator.BuildExternalClient()
		swarmClientSet := k8s.BuildSwarmExternalClient()
		api := operator.BuildAPIExternalClient()

		m := crdop.NewManager(api)
		mng := crd.NewManager(m)
		acc, err := mng.IsAccepted(context.Background())
		if err != nil {
			log.Fatalf("unable to check swarm crd status, error %v", err)
		}
		if !acc {
			if err := mng.Create(context.Background()); err != nil {
				log.Fatalf("unable to initialize swarm crd, error %v", err)
			}
		}

		crdif := crdinformers.NewSharedInformerFactory(swarmClientSet, 0)
		sif := informers.NewSharedInformerFactory(clientSet, 0)

		swi := crdif.K8slab().V1alpha1().Swarms().Informer()
		stsi := sif.Apps().V1().StatefulSets().Informer()
		podi := sif.Core().V1().Pods().Informer()

		crdif.Start(ctx.Done())
		sif.Start(ctx.Done())

		if !cache.WaitForNamedCacheSync(v1alpha1.CrdKind, ctx.Done(), podi.HasSynced, swi.HasSynced, stsi.HasSynced) {
			log.Fatal("unable to sync pod informer")
		}

		swl := crdif.K8slab().V1alpha1().Swarms().Lister()
		stsl := sif.Apps().V1().StatefulSets().Lister()
		podl := sif.Core().V1().Pods().Lister()

		cmp := configmap.NewProvider(clientSet)
		ex := app.NewExecutor(cmp, nil) // @TODO: Refresher
		appm := app.NewManager(ex, swl)
		selSt := statefulset.NewSelectorStore()
		ctl := app.NewSwarmController(swarmClientSet, swl, stsl, podl, selSt, appm) // @TODO: Decompose
		go ctl.Run(ctx)

		crdh := crd.NewHandler(ctl)
		swCtl := operator.New(crdh, swi, v1alpha1.CrdKind)
		go swCtl.Run(ctx)

		stsh := statefulset.NewHandler(ctl, selSt)
		stsCtl := operator.New(stsh, stsi, "StatefulSet")
		go stsCtl.Run(ctx)

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
		cancel()
		_ = srv.Close()

		log.Info("Stopping controller")
	},
}

func init() {
	rootCmd.AddCommand(externalCmd)
}
