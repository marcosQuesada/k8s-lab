# builder stage
FROM golang:1.17.5 as builder
ARG DATE
ARG COMMIT
ARG SERVICE
WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN echo "building k8s lab service: ${SERVICE} on commit: ${COMMIT} date: ${DATE}"
RUN CGO_ENABLED=0 go build -ldflags "-X github.com/marcosQuesada/k8s-lab/config.Commit=${COMMIT} -X github.com/marcosQuesada/k8s-lab/config.Date=${DATE}" ./services/${SERVICE}

# Conditional copy workaround if config is present
RUN mkdir -p /tmp/build/config
RUN cp /app/${SERVICE} /tmp/build
RUN if test -f "/app/services/${SERVICE}/config/config.yml" ; then cp /app/services/${SERVICE}/config/* /tmp/build/config; fi
RUN ls -l /tmp/build/config

# final stage
FROM alpine:3.11.5
ARG SERVICE
COPY --from=builder /tmp/build/* /app/
