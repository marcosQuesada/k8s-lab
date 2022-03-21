# Pool Config Controller

## Docker build
From project root:

```
docker build -t controller . --build-arg SERVICE=config-reloader-controller --build-arg COMMIT=$(git rev-list -1 HEAD) --build-arg DATE=$(date +%m-%d-%Y)
```