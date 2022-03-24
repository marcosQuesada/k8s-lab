# ConfigMap owner controller

Marks owner reference on used configmaps, adding deployment/Statefulset owner parent.

## Project goal

Easy to find configmap relations. Able to find Oprhan configmaps.

Behaviours:
- namespace scope
- all-namespaces scope

## Flow notes
- Controller watches configmaps in all/one namespace
- on create/update looks for deployemnt/statefulsets on configmap namespace until finding the parent
- once found patches configmap ownerReference 

## Corner case
- What to do on orphan configmap
  - It can be marked as Orphan ?

## CRD api generation
```
vendor/k8s.io/code-generator/generate-groups.sh all github.com/marcosQuesada/k8s-lab/services/configmap-claim-owner-controller/internal/infra/k8s/crd/generated github.com/marcosQuesada/k8s-lab/services/configmap-claim-owner-controller/internal/infra/k8s/crd/apis "configmapownerclaim:v1alpha1" --go-header-file ./hack/boilerplate.go.txt --output-base "$(dirname "${BASH_SOURCE[0]}")/" -v 10 
```