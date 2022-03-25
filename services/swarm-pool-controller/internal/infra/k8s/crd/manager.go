package crd

import (
	"context"
	"github.com/marcosQuesada/k8s-lab/pkg/operator/crd"
	"github.com/marcosQuesada/k8s-lab/services/swarm-pool-controller/internal/infra/k8s/crd/apis/swarm/v1alpha1"
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
										"namespace":          {Type: "string"},
										"watched-label":      {Type: "string"},
										"config-map-subject": {Type: "string"},
										"version":            {Type: "integer"},
										"size":               {Type: "integer"},
										"workload": {
											Type: "array",
											Items: &v1.JSONSchemaPropsOrArray{
												Schema: &v1.JSONSchemaProps{
													Type: "string",
												},
											},
										},
										"members": {
											Type: "array",
											Items: &v1.JSONSchemaPropsOrArray{
												Schema: &v1.JSONSchemaProps{
													Type: "object",
													Properties: map[string]v1.JSONSchemaProps{
														"name": {Type: "string"},
														"jobs": {
															Type: "array",
															Items: &v1.JSONSchemaPropsOrArray{
																Schema: &v1.JSONSchemaProps{
																	Type: "string",
																},
															},
														},
														"state": {
															Type: "object",
															Properties: map[string]v1.JSONSchemaProps{
																"phase": {
																	Type: "string",
																},
															},
														},
														"created_at": {Type: "integer"},
													},
												},
											},
											Required: []string{"created_at"},
										},
									},
									Required: []string{"namespace", "watched-label", "workload"},
								},
								"status": {
									Type: "object",
									Properties: map[string]v1.JSONSchemaProps{
										"phase": {
											Type: "string",
										},
									},
								},
							},
						},
					},
					AdditionalPrinterColumns: []v1.CustomResourceColumnDefinition{
						{
							Name:     "Namespace",
							Type:     "string",
							JSONPath: ".spec.namespace",
						},
						{
							Name:     "WatchedLabel",
							Type:     "string",
							JSONPath: ".spec.watched-label",
						},
						{
							Name:     "Version",
							Type:     "integer",
							JSONPath: ".spec.version",
						},
						{
							Name:     "Size",
							Type:     "integer",
							JSONPath: ".spec.size",
						},
						{
							Name:     "Age",
							Type:     "date",
							JSONPath: ".metadata.creationTimestamp",
						},
						{
							Name:     "Status",
							Type:     "string",
							JSONPath: ".status.phase",
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
