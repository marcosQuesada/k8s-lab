package crd

import (
	"context"
	"github.com/davecgh/go-spew/spew"
	"github.com/marcosQuesada/k8s-lab/pkg/operator"
	"github.com/marcosQuesada/k8s-lab/pkg/operator/crd"
	"testing"
)

func TestItRecognizedCreatedCrdDevelopment(t *testing.T) {
	api := operator.BuildAPIExternalClient()
	i := crd.NewManager(api)
	m := NewManager(i)

	err := m.Create(context.Background())

	spew.Dump(err)
}

func TestItChecksCRDAcceptedDevelopment(t *testing.T) {
	api := operator.BuildAPIExternalClient()
	i := crd.NewManager(api)
	m := NewManager(i)

	e, err := m.IsAccepted(context.Background())

	spew.Dump(e, err)
}
