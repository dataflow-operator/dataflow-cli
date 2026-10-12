/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	dataflowv1 "github.com/dataflow-operator/dataflow/api/v1"
	runpkg "github.com/dataflow-operator/dataflow/pkg/run"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunOnce_KafkaRejected(t *testing.T) {
	path := writeManifest(t, `
apiVersion: dataflow.dataflow.io/v1
kind: DataFlow
metadata:
  name: stream
spec:
  source:
    type: kafka
  sink:
    type: postgresql
`)
	restore := snapshotRunFlags()
	t.Cleanup(restore)
	runOnce = true
	runLogLevel = "error"

	err := runRun(nil, []string{path})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "does not finish a one-shot run")
	assert.Contains(t, err.Error(), "kafka")
}

func TestRunOnce_PrintsSummaryWhenSourceExhausts(t *testing.T) {
	path := writeManifest(t, `
apiVersion: dataflow.dataflow.io/v1
kind: DataFlowCron
metadata:
  name: pg-migration
  namespace: etl
spec:
  schedule: "0 2 * * *"
  source:
    type: postgresql
  sink:
    type: postgresql
  triggers:
    - name: validate
      image: postgres:16
`)
	restore := snapshotRunFlags()
	t.Cleanup(restore)
	runOnce = true
	runReset = true
	runTimeout = time.Minute
	runLogLevel = "error"

	orig := executePipeline
	t.Cleanup(func() { executePipeline = orig })
	executePipeline = func(ctx context.Context, spec *dataflowv1.DataFlowSpec, opts runpkg.RunOptions) (runpkg.RunResult, error) {
		deadline, ok := ctx.Deadline()
		assert.True(t, ok)
		assert.WithinDuration(t, time.Now().Add(time.Minute), deadline, 2*time.Second)
		assert.True(t, opts.Once)
		assert.True(t, opts.ResetCheckpoint)
		assert.Equal(t, "etl", opts.Namespace)
		assert.Equal(t, "pg-migration", opts.Name)
		assert.Equal(t, filepath.Join(".dataflow", "pg-migration.json"), opts.CheckpointPath)
		assert.Equal(t, "postgresql", spec.Source.Type)
		return runpkg.RunResult{Processed: 4, Errors: 1, CheckpointPath: opts.CheckpointPath}, nil
	}

	var out, warn bytes.Buffer
	summaryOut = &out
	warnOut = &warn
	t.Cleanup(func() {
		summaryOut = os.Stdout
		warnOut = os.Stderr
	})

	err := runRun(nil, []string{path})
	require.NoError(t, err)
	assert.Contains(t, warn.String(), "DataFlowCron schedule and triggers are ignored")
	assert.Contains(t, out.String(), "source: postgresql")
	assert.Contains(t, out.String(), "sink: postgresql")
	assert.Contains(t, out.String(), "processed: 4")
	assert.Contains(t, out.String(), "errors: 1")
	assert.Contains(t, out.String(), "checkpoint: .dataflow/pg-migration.json")
}

func TestRunResetRequiresCheckpoint(t *testing.T) {
	path := writeManifest(t, `
apiVersion: dataflow.dataflow.io/v1
kind: DataFlow
metadata:
  name: stream
spec:
  source:
    type: kafka
  sink:
    type: postgresql
`)
	restore := snapshotRunFlags()
	t.Cleanup(restore)
	runReset = true
	runOnce = false
	runCheckpointFile = ""

	err := runRun(nil, []string{path})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--reset requires --once or --checkpoint-file")
}

func writeManifest(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "manifest.yaml")
	require.NoError(t, os.WriteFile(path, []byte(body), 0o644))
	return path
}

func snapshotRunFlags() func() {
	logLevel, once, file, reset, timeout := runLogLevel, runOnce, runCheckpointFile, runReset, runTimeout
	return func() {
		runLogLevel, runOnce, runCheckpointFile, runReset, runTimeout = logLevel, once, file, reset, timeout
	}
}
