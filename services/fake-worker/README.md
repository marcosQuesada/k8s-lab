# Fake Worker
Sample application loading config from file supporting hot reload.

## Docker build
From project root:

```
docker build -t fake-worker . --build-arg SERVICE=fake-worker --build-arg COMMIT=$(git rev-list -1 HEAD) --build-arg DATE=$(date +%m-%d-%Y)
```