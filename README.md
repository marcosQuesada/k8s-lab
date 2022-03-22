# K8s-lab [WIP]

Monorepo dedicated to easy boostrap Kubernetes operators, controllers & CRDs. It's a kubernetes lab playground allowing dig on kubernetes internals.

## Projects

- **config-reloader-controller**
  - watches concretes configmaps, on change detected refresh deployment/statefulset ensuring all pods are running in the last updated config version.
- **swarm-pool-controller**
  - shards jobs to a workers pool in an ordered deterministic way
  - Those jobs cannot be executed concurrently, so that, it balances jobs to the workers pool, allowing scale up/down in a transparent way.
- **scheduler-controler**
  - schedule actions (shell commands) in ticker/timer fashion (redoing what cronjob does indeed)

## Minikube setup

The whole system can be tested using minikube, a few notes to make it work:

Start minikube cluster:
```
minikube start --addons=ingress --cpus=2 --cni=flannel --install-addons=true --kubernetes-version=stable --memory=6g --driver=docker
```

Stop minikube cluster:
```
minikube stop
```

### Build docker image
Bear in mind that minikube images are expected to be local, if you check k8s manifest you will see that imagePullPolicy is Never in all the cases. [This will be fixed using Kustomize in further iterations]

Before building images ensure minikube is pointing docker env:
```
eval $(minikube docker-env)
```
