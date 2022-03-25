package k8s

import (
	"github.com/marcosQuesada/k8s-lab/services/config-reloader-controller/internal/infra/k8s/crd/generated/clientset/versioned"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
)

// BuildConfigMapPodRefresherInternalClient instantiates internal client
func BuildConfigMapPodRefresherInternalClient() versioned.Interface {
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("unable to get In cluster config, error %v", err)
	}

	client, err := versioned.NewForConfig(config)
	if err != nil {
		log.Fatalf("unable to build client from config, error %v", err)
	}

	return client
}

// BuildConfigMapPodRefresherExternalClient instantiates local client with user credentials
func BuildConfigMapPodRefresherExternalClient() versioned.Interface {
	kubeConfigPath := os.Getenv("HOME") + "/.kube/config"

	config, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		log.Fatalf("unable to get cluster config from flags, error %v", err)
	}

	client, err := versioned.NewForConfig(config)
	if err != nil {
		log.Fatalf("unable to build client from config, error %v", err)
	}

	return client
}
