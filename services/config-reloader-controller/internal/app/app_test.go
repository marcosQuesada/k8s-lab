package app

import (
	"context"
	"github.com/davecgh/go-spew/spew"
	"github.com/marcosQuesada/k8s-lab/pkg/operator"
	"github.com/marcosQuesada/k8s-lab/services/config-reloader-controller/internal/infra/k8s/crd/apis/configmappodrefresher/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestItPopulatesRefreshersFromCreatedEvent(t *testing.T) {
	watchedCmNamespacee := "default"
	watchedConfigMap := "fakeconfigmap"
	//deployName := "fakedeployment"

	a := NewApp()
	c := fakeConfigMapRefresher("default", "foo", watchedCmNamespacee, watchedConfigMap)

	a.Watch(c)

	if !a.IsRegistered(watchedCmNamespacee, watchedConfigMap) {
		t.Fatalf("unable to find watched configmap %s on namespace %s", watchedConfigMap, watchedCmNamespacee)
	}
}

func fakeConfigMapRefresher(namespace, name, c, wcm string) *v1alpha1.ConfigMapPodRefresher {
	return &v1alpha1.ConfigMapPodRefresher{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMapPodRefresher",
			APIVersion: "v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1alpha1.ConfigMapPodsRefresherSpec{
			Version:          0,
			Namespace:        c,
			WatchedConfigMap: wcm,
			PoolType:         v1alpha1.Deployment,
			PoolSubjectName:  "fakedeployment",
		},
		Status: v1alpha1.PoolStatus{},
	}
}

func TestWatchAllConfigMapsFromAllNamespaces(t *testing.T) {
	clientset := operator.BuildExternalClient()

	cms, err := clientset.CoreV1().ConfigMaps("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	spew.Dump(cms)

	dp, err := clientset.AppsV1().Deployments("").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	spew.Dump(dp)
	//controllerRef := metav1.GetControllerOf(pod)
}
