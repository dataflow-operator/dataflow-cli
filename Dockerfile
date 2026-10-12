# syntax=docker/dockerfile:1
# Context is the dataflow-cli module root:
#   docker build -t dataflow-cli .
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG TARGETARCH
ARG VERSION=dev
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} \
    go build -trimpath -ldflags "-s -w -X github.com/dataflow-operator/dataflow-cli/cmd.Version=${VERSION}" \
    -o /out/dataflow-cli .

FROM alpine:3.21

RUN apk --no-cache add ca-certificates
WORKDIR /work

COPY --from=builder /out/dataflow-cli /usr/local/bin/dataflow-cli

ENTRYPOINT ["dataflow-cli"]
