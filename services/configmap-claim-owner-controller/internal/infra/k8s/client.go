package k8s

import (
	"github.com/marcosQuesada/k8s-lab/services/configmap-claim-owner-controller/internal/infra/k8s/crd/generated/clientset/versioned"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
)

// BuildConfigMapClaimOwnerInternalClient instantiates internal client
func BuildConfigMapClaimOwnerInternalClient() versioned.Interface {
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

// BuildConfigMapClaimOwnerExternalClient instantiates local client with user credentials
func BuildConfigMapClaimOwnerExternalClient() versioned.Interface {
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
