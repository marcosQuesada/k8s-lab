package crd

import (
	"context"
	"github.com/davecgh/go-spew/spew"
	"github.com/marcosQuesada/k8s-lab/pkg/operator"
	"testing"
)

func TestItRecognizedCreatedCrdDevelopment(t *testing.T) {
	api := operator.BuildAPIExternalClient()
	m := NewManager(api)

	e, err := m.IsAccepted(context.Background(), "swarms.k8slab.info")

	spew.Dump(e, err)
}
