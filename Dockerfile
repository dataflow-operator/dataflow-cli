# syntax=docker/dockerfile:1
# Build context is the monorepo root (the CLI replaces ../dataflow):
#   docker build -f dataflow-cli/Dockerfile -t dataflow-cli .
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /src
COPY dataflow/go.mod dataflow/go.sum /src/dataflow/
COPY dataflow-cli/go.mod dataflow-cli/go.sum /src/dataflow-cli/

WORKDIR /src/dataflow-cli
ENV GOWORK=off
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY dataflow /src/dataflow
COPY dataflow-cli /src/dataflow-cli

ARG TARGETARCH
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} \
    go build -trimpath -ldflags "-s -w" \
    -o /out/dataflow-cli .

FROM alpine:3.21

RUN apk --no-cache add ca-certificates
WORKDIR /work

COPY --from=builder /out/dataflow-cli /usr/local/bin/dataflow-cli

ENTRYPOINT ["dataflow-cli"]
