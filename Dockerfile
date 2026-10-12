# syntax=docker/dockerfile:1
# Build context is the monorepo root. go.work selects the local dataflow module.
#   docker build -f dataflow-cli/Dockerfile -t dataflow-cli .
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /src
COPY go.work go.work.sum ./
COPY dataflow/go.mod dataflow/go.sum ./dataflow/
COPY dataflow-cli/go.mod dataflow-cli/go.sum ./dataflow-cli/
RUN cd dataflow-cli && go mod download

COPY dataflow ./dataflow
COPY dataflow-cli ./dataflow-cli

ARG TARGETARCH
ARG VERSION=dev
RUN cd dataflow-cli && CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} \
    go build -trimpath -ldflags "-s -w -X github.com/dataflow-operator/dataflow-cli/cmd.Version=${VERSION}" \
    -o /out/dataflow-cli .

FROM alpine:3.21

RUN apk --no-cache add ca-certificates
WORKDIR /work

COPY --from=builder /out/dataflow-cli /usr/local/bin/dataflow-cli

ENTRYPOINT ["dataflow-cli"]
