# Swarm Pool Controller

Swarm Pool Controller solves scenarios were Collaborative Workload consumption is required, scenarios  where parallel job processing does not increase overall performance.

They are special cases as they are not the typical deployment scenarios, but there are many scenarios like that, from stream consuming where packets cannot be marked as processing (in-flight), so that, multiple consumers will just reprocess the same source.

As example, dedicated streams as real time video, long run jobs (database backups) and many others.

## Controller flow

- Controller watches Swarm CRDs that defines Statefulset/Pod watched labels
- Watched pools create an ordered pool, detects pool completion once Statefulset size matches pool size
- On Pool size change balances workload jobs between workers through a concrete configmap

## Current features

- Controller workload definition trough watched CRD, on create/update/delete balance workload jobs on workers pool
- Scaling Up/Down balances workload assignations on the updated workers pool

### Minikube deploy
- Apply required manifests (in order), namespace, rbac, configmaps, operator and statefulset.
```
kubectl get pods -w

NAME                                READY   STATUS    RESTARTS   AGE
swarm-controller-7bcc789689-sj628   1/1     Running   0          10h
swarm-worker-0                      1/1     Running   0          10h
swarm-worker-1                      1/1     Running   0          10h

```
Check worker assignations from worker config:
```
kubectl describe cm swarm-worker-config 
Namespace:    swarm
Labels:       <none>
Annotations:  <none>

Data
====
config.yml:
----
workloads:
  swarm-worker-0:
    jobs:
      - stream:xxrtve1
      - stream:xxrtve2
      - stream:zrtve2:new
      - stream:zrtve1:new
      - stream:zrtve0:new
      - stream:cctv0:updated
      - stream:history:new
      - stream:foo:new
      - stream:xxctv3:new
      - stream:cctv3:updated
      - stream:xxctv0:updated
      - stream:xxctv10:updated
      - stream:xxctv11:updated
      - stream:xxctv12:updated
    version: 0
  swarm-worker-1:
    jobs:
      - stream:xxctv13:updated
      - stream:xxctv14:updated
      - stream:yxctv1:updated
      - stream:yxctv2:updated
      - stream:yxctv3:updated
      - stream:xabcn0:updated
      - stream:xacb01:updated
      - stream:xacb02:updated
      - stream:xacb03:updated
      - stream:xacb04:updated
      - stream:sportnews0:updated
      - stream:cars:new
      - stream:cars:updated 
```

You can even introspect what are the configs mounted as volumes in the worker pods:
```
kubectl exec -ti swarm-worker-0 -- cat /app/config/config.yml
```
Scaling Up/Down
```
`kuectl scale statefulset swarm-worker --replicas=3 -n swarm 
```
On Scaling Up Operator will pick up pool changes, balance workloads over the new updated workers pool, updated workers shared config (swarm-worker-config configmap), it will schedule to reboot worker0 and 1, so that, after refresh cycle all workers will belong to the same config version.

## Docker build
From project root:

```
docker build -t controller . --build-arg SERVICE=config-reloader-controller --build-arg COMMIT=$(git rev-list -1 HEAD) --build-arg DATE=$(date +%m-%d-%Y)
```


## Minikube rollout
```
docker tag swarm-controller:latest swarm-controller:a4c0d90019d8
```
```
kubectl set image deployment/swarm-controller swarm-controller=swarm-controller:01a68000eb00 -n swarm
kubectl set image statefulset/swarm-worker swarm-worker=swarm-worker:2ddc673264ed -n swarm
```
```
kubectl rollout restart deployment/swarm-controller
kubectl rollout restart statefulset/swarm-worker
```

## Development Notes

### CRD API generation
```
vendor/k8s.io/code-generator/generate-groups.sh all github.com/marcosQuesada/k8s-lab/services/swarm-pool-controller/internal/infra/k8s/crd/generated github.com/marcosQuesada/k8s-lab/services/swarm-pool-controller/internal/infra/k8s/crd/apis "swarm:v1alpha1" --go-header-file ./hack/boilerplate.go.txt --output-base "$(dirname "${BASH_SOURCE[0]}")/" -v 10 
```