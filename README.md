# dataflow-cli

CLI for the DataFlow operator: the same Web API actions as the web UI, plus local manifest runs without Kubernetes.

Documentation: [CLI (docs)](https://dataflow-operator.github.io/docs/cli/).

## Install

### Homebrew

```bash
brew install dataflow-operator/tap/dataflow-cli
```

The formula is published to the `dataflow-operator/tap` tap with each `v*` release (macOS and Linux, amd64 and arm64).

### Release binary

A `v*` tag publishes binaries and `SHA256SUMS` to [GitHub Releases](https://github.com/dataflow-operator/dataflow-cli/releases).

| OS | Architecture | Asset |
|----|--------------|-------|
| Linux | amd64 | `dataflow-cli-linux-amd64` |
| Linux | arm64 | `dataflow-cli-linux-arm64` |
| Linux | armv7 | `dataflow-cli-linux-arm-v7` |
| macOS | amd64 | `dataflow-cli-darwin-amd64` |
| macOS | arm64 | `dataflow-cli-darwin-arm64` |
| Windows | amd64 | `dataflow-cli-windows-amd64.exe` |
| Windows | arm64 | `dataflow-cli-windows-arm64.exe` |

```bash
curl -fsSL -o dataflow-cli \
  "https://github.com/dataflow-operator/dataflow-cli/releases/download/vX.Y.Z/dataflow-cli-linux-amd64"
chmod +x dataflow-cli
./dataflow-cli --help
```

Replace `vX.Y.Z` and the asset name for your OS.

### Docker

Images are published to `ghcr.io/dataflow-operator/dataflow-cli` (`linux/amd64` and `linux/arm64`). `:latest` and the version tag come from a `v*` release. `:main` is updated from the `main` branch.

```bash
docker pull ghcr.io/dataflow-operator/dataflow-cli:latest

docker run --rm ghcr.io/dataflow-operator/dataflow-cli:latest --help

# Local manifest. On Linux, --network host reaches source/sink on the host.
# On Docker Desktop, use host.docker.internal in the manifest instead of localhost.
docker run --rm --network host \
  -v "$PWD:/work" -w /work \
  ghcr.io/dataflow-operator/dataflow-cli:latest \
  run manifest.yaml --once

# SecretRef is read from the mounted kubeconfig.
docker run --rm --network host \
  -v "$HOME/.kube:/root/.kube:ro" \
  -v "$PWD:/work" -w /work \
  ghcr.io/dataflow-operator/dataflow-cli:latest \
  run manifest.yaml --once
```

Build from the CLI module directory (`dataflow-cli/` in the monorepo, the repository root in `dataflow-operator/dataflow-cli`):

```bash
docker build -t dataflow-cli .
```

`task docker` runs the same build.

### From source

- `go build -o dataflow-cli .` in `dataflow-cli` ([Go 1.25+](https://go.dev/dl/)).
- `go.mod` requires `github.com/dataflow-operator/dataflow` at `v1.12.1-0.20261012061201-dec4eee83b3d` (branch `feat/pkg-run-oneshot`). The tag `v1.12.0` does not include one-shot `pkg/run` options. There is no local `replace`.
- In the monorepo, `go.work` still builds against the local `dataflow` checkout.
- `dataflow-cli --version` prints `dev` unless the build sets `-ldflags "-X github.com/dataflow-operator/dataflow-cli/cmd.Version=..."`. `task build` fills that from `git describe`.

A push to `main` of the `dataflow-operator/dataflow-cli` repository runs semantic-release. It tags `vX.Y.Z` and moves the `dataflow` requirement to `@latest` only when that module still builds the CLI. The same workflow stamps that version into the binaries and the image. The GitHub token that creates the tag does not start a second workflow.

## Web API (same actions as the web UI)

The default base URL is `http://localhost:8080/api`. Override it with `--api-url` or the `DATAFLOW_API_URL` environment variable.

```bash
# Namespaces
dataflow-cli namespaces

# DataFlows
dataflow-cli dataflows list
dataflow-cli dataflows list --namespace=my-ns
dataflow-cli dataflows get my-dataflow
dataflow-cli dataflows create --file=manifest.yaml
dataflow-cli dataflows update my-dataflow --file=manifest.yaml
dataflow-cli dataflows delete my-dataflow

# Logs (follow the stream with --follow)
dataflow-cli logs my-dataflow --tail=200
dataflow-cli logs my-dataflow -f

# Status and metrics
dataflow-cli status my-dataflow
dataflow-cli metrics my-dataflow
```

Every command accepts `--namespace` (default `default`).

## Local manifest run (no Kubernetes)

The pipeline runs inside the CLI process. The cluster and the controller are not required. Source and sink must be reachable (Kafka, PostgreSQL, and so on).

```bash
dataflow-cli run manifest.yaml
dataflow-cli run manifest.yaml --log-level=debug
```

Stop with Ctrl+C.

## One-shot migration

`--once` reads a polling source until it is exhausted, flushes the sink, writes a local checkpoint, and exits. Use it for a one-time snapshot (for example PostgreSQL to PostgreSQL) without a cluster. `kind: DataFlowCron` is accepted; `schedule` and `triggers` are ignored.

Supported sources: `postgresql`, `clickhouse`, `trino`, `iceberg`, `nessie`. `kafka` and `postgresql-cdc` never close, so `--once` fails before connecting.

```bash
dataflow-cli run manifest.yaml --once
dataflow-cli run manifest.yaml --once --checkpoint-file .dataflow/pg-migration.json --reset
dataflow-cli run manifest.yaml --once --timeout 2h
```

| Flag | Description |
|------|-------------|
| `--once` | Exit when the source is exhausted |
| `--checkpoint-file` | Checkpoint JSON. Default for `--once`: `.dataflow/<metadata.name>.json` |
| `--reset` | Clear that source's checkpoint before start |
| `--timeout` | Maximum run time (`2h`, `30m`). `0` means no limit |
| `--log-level` | `debug`, `info`, `warn`, `error` |

Ctrl+C or a timeout flushes the checkpoint to disk and exits with a non-zero status. The next run resumes from the file. On success the CLI prints source, sink, processed, errors, and the checkpoint path.

`checkpointPersistence: false` skips the file. `SecretRef` is read from the current kubeconfig (or in-cluster config) before the pipeline starts. A missing secret or key stops the run before connect.

The manifest stays a normal DataFlow. On the source set `changeTrackingColumn`, `orderByColumn`, and `readBatchSize`. On the sink set `upsertMode` and `conflictKey`.

## Example manifest

Standard DataFlow CR (YAML):

```yaml
apiVersion: dataflow.dataflow.io/v1
kind: DataFlow
metadata:
  name: kafka-to-postgres
spec:
  source:
    type: kafka
    config:
      brokers:
        - localhost:9092
      topic: input-topic
      consumerGroup: dataflow-group
  sink:
    type: postgresql
    config:
      connectionString: "postgres://user:pass@localhost:5432/db?sslmode=disable"
      table: output_table
```

## CI

In the published repository the workflow is [`.github/workflows/build.yml`](.github/workflows/build.yml). It runs on every push to `main` and on pull requests. There is no path filter.

- tests on pull requests and pushes to `main`;
- semantic-release on `main`: tag `vX.Y.Z`, and `go get @latest` only if that `dataflow` module still builds the CLI;
- binaries for Linux (amd64, arm64, armv7), macOS (amd64, arm64), and Windows (amd64, arm64), with the version linked into the binary;
- container image `ghcr.io/dataflow-operator/dataflow-cli` for `linux/amd64` and `linux/arm64`;
- GitHub Release assets attached in the same run that creates the `v*` tag.
