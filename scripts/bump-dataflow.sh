#!/bin/sh
# Point go.mod at the newest dataflow release that still builds this CLI.
# v1.12.0 does not have pkg/run one-shot options; do not replace a working require with it.
set -eu

mod=github.com/dataflow-operator/dataflow
current=$(GOWORK=off go list -m -f '{{.Version}}' "$mod")

GOWORK=off go get "${mod}@latest"

if ! GOWORK=off go test -c -o /tmp/dataflow-cli-bump.test .; then
  echo "dataflow@latest does not build the CLI; keeping ${current}" >&2
  GOWORK=off go get "${mod}@${current}"
fi

rm -f /tmp/dataflow-cli-bump.test
GOWORK=off go mod tidy
