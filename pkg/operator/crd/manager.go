package crd

import (
	"context"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
	"time"
)

const pollFrequency = time.Second
const pollTimeout = 10 * time.Second

type Initializer interface {
	Create(ctx context.Context, cr *v1.CustomResourceDefinition) error
	IsAccepted(ctx context.Context, resourceName string) (bool, error)
}

type manager struct {
	apiExtensionsClientSet apiextensionsclientset.Interface
}

func NewManager(api apiextensionsclientset.Interface) Initializer {
	return &manager{apiExtensionsClientSet: api}
}

func (c *manager) Create(ctx context.Context, cr *v1.CustomResourceDefinition) error {
	_, err := c.apiExtensionsClientSet.ApiextensionsV1().CustomResourceDefinitions().Create(ctx, cr, metav1.CreateOptions{})

	if err != nil {
		klog.Fatalf("unable to create CRD , error: %v", err.Error())
	}

	return c.waitCRDAccepted(ctx, cr.Name)
}

func (c *manager) IsAccepted(ctx context.Context, resourceName string) (bool, error) {
	cr, err := c.apiExtensionsClientSet.ApiextensionsV1().CustomResourceDefinitions().Get(ctx, resourceName, metav1.GetOptions{})
	if err != nil {
		return false, err
	}

	spew.Dump(cr.Status.Conditions)
	for _, condition := range cr.Status.Conditions {
		if condition.Type == v1.Established &&
			condition.Status == v1.ConditionTrue {
			return true, nil
		}
	}

	return false, fmt.Errorf("CRD is not accepted")
}

func (c *manager) waitCRDAccepted(ctx context.Context, resourceName string) error {
	err := wait.Poll(pollFrequency, pollTimeout, func() (bool, error) {
		return c.IsAccepted(ctx, resourceName)
	})

	return err
}
