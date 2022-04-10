package pod

import (
	"context"
	"github.com/davecgh/go-spew/spew"
	"github.com/marcosQuesada/k8s-lab/pkg/operator"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"testing"
)

func TestItGetsWorkloadConfigFromPodUsingExec(t *testing.T) {
	clientSet := operator.BuildExternalClient()
	kubeConfigPath := os.Getenv("HOME") + "/.kube/config"
	restConfig, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		t.Fatalf("unable to get cluster config from flags, error %v", err)
	}

	p := NewProvider(clientSet, restConfig)
	namespace := "swarm"
	name := "swarm-worker-0"
	w, err := p.Workload(context.Background(), namespace, name)
	if err != nil {
		t.Fatalf("unable to decode workload, error %v", err)
	}

	spew.Dump(w)
}
