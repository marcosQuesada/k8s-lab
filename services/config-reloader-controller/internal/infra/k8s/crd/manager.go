package crd

import (
	"context"
	"github.com/marcosQuesada/k8s-lab/pkg/operator/crd"
	"github.com/marcosQuesada/k8s-lab/services/config-reloader-controller/internal/infra/k8s/crd/apis/configmappodrefresher/v1alpha1"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type manager struct {
	initializer crd.Initializer
}

func NewManager(i crd.Initializer) *manager {
	return &manager{
		initializer: i,
	}
}

func (m *manager) Create(ctx context.Context) error {
	cr := &v1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: v1alpha1.Name,
		},
		Spec: v1.CustomResourceDefinitionSpec{
			Group: crd.GroupName,
			Versions: []v1.CustomResourceDefinitionVersion{
				{
					Name:    v1alpha1.Version,
					Served:  true,
					Storage: true,
					Subresources: &v1.CustomResourceSubresources{
						Status: &v1.CustomResourceSubresourceStatus{},
					},
					Schema: &v1.CustomResourceValidation{
						OpenAPIV3Schema: &v1.JSONSchemaProps{
							Type: "object",
							Properties: map[string]v1.JSONSchemaProps{
								"spec": {
									Type: "object",
									Properties: map[string]v1.JSONSchemaProps{
										"version":            {Type: "integer"},
										"namespace":          {Type: "string"},
										"watched-config-map": {Type: "string"},
										"pool-type":          {Type: "string"},
										"pool-subject-name":  {Type: "string"},
									},
									Required: []string{"namespace", "watched-config-map", "pool-type", "pool-subject-name"},
								},
							},
						},
					},
					AdditionalPrinterColumns: []v1.CustomResourceColumnDefinition{
						{
							Name:     "Version",
							Type:     "string",
							JSONPath: ".spec.version",
						},
						{
							Name:     "Namespace",
							Type:     "string",
							JSONPath: ".spec.namespace",
						},
						{
							Name:     "WatchedConfigMap",
							Type:     "string",
							JSONPath: ".spec.watched-config-map",
						},
						{
							Name:     "PoolSubjectName",
							Type:     "string",
							JSONPath: ".spec.pool-subject-name",
						},
					},
				},
			},
			Scope: v1.NamespaceScoped,
			Names: v1.CustomResourceDefinitionNames{
				Plural:     v1alpha1.Plural,
				Singular:   v1alpha1.Singular,
				Kind:       v1alpha1.CrdKind,
				ShortNames: []string{v1alpha1.ShortName},
			},
		},
	}

	return m.initializer.Create(ctx, cr)
}

func (m *manager) IsAccepted(ctx context.Context) (bool, error) {
	return m.initializer.IsAccepted(ctx, v1alpha1.Name)
}
