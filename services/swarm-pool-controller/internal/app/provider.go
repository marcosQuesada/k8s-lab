package app

import (
	"fmt"
	"github.com/marcosQuesada/k8s-lab/pkg/operator/pod"
	swapi "github.com/marcosQuesada/k8s-lab/services/swarm-pool-controller/internal/infra/k8s/crd/apis/swarm/v1alpha1"
	"github.com/marcosQuesada/k8s-lab/services/swarm-pool-controller/internal/infra/k8s/crd/generated/listers/swarm/v1alpha1"
	log "github.com/sirupsen/logrus"
	api "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	appv1 "k8s.io/client-go/listers/apps/v1"
	v1 "k8s.io/client-go/listers/core/v1"
)

type provider struct {
	swarmLister       v1alpha1.SwarmLister
	statefulSetLister appv1.StatefulSetLister
	podLister         v1.PodLister
}

func NewProvider(swl v1alpha1.SwarmLister, stsl appv1.StatefulSetLister, pl v1.PodLister) *provider {
	return &provider{
		swarmLister:       swl,
		statefulSetLister: stsl,
		podLister:         pl,
	}
}
func (c *provider) PodNamesFromSelector(namespace string, ls *metav1.LabelSelector) ([]string, error) {
	selector, err := metav1.LabelSelectorAsSelector(ls)
	if err != nil {
		return nil, fmt.Errorf("unable to get label selector, error %v", err)
	}

	pods, err := c.podLister.Pods(namespace).List(selector)
	if err != nil {
		return nil, fmt.Errorf("unable to get pods from selector, error %v", err)
	}

	var names []string
	for _, pd := range pods {
		if pod.IsTerminated(pd) || pod.HasDeletionTimestamp(pd) {
			continue
		}
		names = append(names, pd.Name)
	}

	return names, nil
}
func (c *provider) StatefulSet(namespace, name string) (*api.StatefulSet, error) {
	sts, err := c.statefulSetLister.StatefulSets(namespace).Get(name)
	if err != nil {
		return nil, fmt.Errorf("unable to get statefulset on namespace %s name %s error %v", name, namespace, err)
	}
	return sts, nil
}

func (c *provider) StatefulSetFromSwarmName(namespace, name string) (*api.StatefulSet, error) {
	sw, err := c.swarmLister.Swarms(namespace).Get(name)
	if err != nil {
		return nil, fmt.Errorf("unable to get swarm namespace %s name %s", namespace, name)
	}

	log.Infof("Processing swarm %s namespace %s statefulset name %s configmap name %s version %d workloads %v",
		sw.Name, sw.Namespace, sw.Spec.StatefulSetName, sw.Spec.ConfigMapName, sw.Spec.Version, sw.Spec.Workload)

	sts, err := c.statefulSetLister.StatefulSets(sw.Namespace).Get(sw.Spec.StatefulSetName)
	if err != nil {
		return nil, fmt.Errorf("unable to get statefulset from swarm %s on namespace %s name %s error %v", sw.Name, sw.Namespace, sw.Spec.StatefulSetName, err)
	}
	return sts, nil
}

func (c *provider) Swarm(namespace, name string) (*swapi.Swarm, error) {
	sw, err := c.swarmLister.Swarms(namespace).Get(name)
	if err != nil {
		return nil, fmt.Errorf("unable to get swarm namespace %s name %s error %v", namespace, name, err)
	}

	return sw, nil
}

func (c *provider) SwarmNameFromStatefulSetName(namespace, name string) (string, error) {
	sws, err := c.swarmLister.Swarms(namespace).List(labels.NewSelector())
	if err != nil {
		return "", fmt.Errorf("unable to list swarm namespace %s name %s", namespace, name)
	}

	for _, sw := range sws {
		if sw.Spec.StatefulSetName == name {
			return sw.Name, nil
		}
	}

	return "", fmt.Errorf("unable to find swarm name from statefulset %s", name)
}
