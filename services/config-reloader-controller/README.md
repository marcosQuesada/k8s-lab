# Config Reloader Controller

It watches concretes configmaps, on change detected refresh deployment/statefulset ensuring all pods are running in the last updated config version.

## Controller flow

- controller watches ConfigMapWatcher CRDs, they define namespace/configmap watch subject and deployment/statefulset to be refreshed on configmap change
- on configmap change detected updates deployment/statefulset to redeploy their pods

## Run external controller [development flow]
```
go run ./services/config-reloader-controller external
```

## Docker build

From project root:
```
docker build -t config-reloader-controller . --build-arg SERVICE=config-reloader-controller --build-arg COMMIT=$(git rev-list -1 HEAD) --build-arg DATE=$(date +%m-%d-%Y)

```

## Development notes

Generate CRD api:
```
vendor/k8s.io/code-generator/generate-groups.sh all github.com/marcosQuesada/k8s-lab/services/config-reloader-controller/internal/infra/k8s/crd/generated github.com/marcosQuesada/k8s-lab/services/config-reloader-controller/internal/infra/k8s/crd/apis "configmappodsrefresher:v1alpha1" --go-header-file ./hack/boilerplate.go.txt --output-base "$(dirname "${BASH_SOURCE[0]}")/" -v 3 

```