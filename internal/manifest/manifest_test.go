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

package manifest

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseDataFlow(t *testing.T) {
	yaml := `
apiVersion: dataflow.dataflow.io/v1
kind: DataFlow
metadata:
  name: test-flow
  namespace: default
spec:
  source:
    type: kafka
    kafka:
      brokers: ["localhost:9092"]
      topic: t
      consumerGroup: g
  sink:
    type: postgresql
    postgresql:
      connectionString: "postgres://localhost/db"
      table: out
`
	df, err := ParseDataFlow([]byte(yaml))
	require.NoError(t, err)
	require.NotNil(t, df)
	assert.Equal(t, "test-flow", df.Name)
	assert.Equal(t, "default", df.Namespace)
	assert.Equal(t, "kafka", df.Spec.Source.Type)
	assert.Equal(t, "postgresql", df.Spec.Sink.Type)
}

func TestLoadDataFlow(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "df.yaml")
	err := os.WriteFile(path, []byte(`
apiVersion: dataflow.dataflow.io/v1
kind: DataFlow
metadata:
  name: from-file
spec:
  source:
    type: kafka
    kafka:
      brokers: ["b:9092"]
      topic: t
      consumerGroup: g
  sink:
    type: postgresql
    postgresql:
      connectionString: "postgres:///"
      table: t
`), 0644)
	require.NoError(t, err)

	df, err := LoadDataFlow(path)
	require.NoError(t, err)
	require.NotNil(t, df)
	assert.Equal(t, "from-file", df.Name)
}

func TestParsePipeline_DataFlowCron(t *testing.T) {
	p, err := ParsePipeline([]byte(`
apiVersion: dataflow.dataflow.io/v1
kind: DataFlowCron
metadata:
  name: pg-migration
  namespace: etl
spec:
  schedule: "0 2 * * *"
  source:
    type: postgresql
    config:
      table: public.src
  sink:
    type: postgresql
    config:
      table: public.dst
  triggers:
    - name: validate
      image: postgres:16
`))
	require.NoError(t, err)
	assert.Equal(t, "DataFlowCron", p.Kind)
	assert.Equal(t, "pg-migration", p.Name)
	assert.Equal(t, "etl", p.Namespace)
	assert.Equal(t, "postgresql", p.Spec.Source.Type)
	assert.Equal(t, "postgresql", p.Spec.Sink.Type)
}

func TestParsePipeline_UnsupportedKind(t *testing.T) {
	_, err := ParsePipeline([]byte(`
apiVersion: dataflow.dataflow.io/v1
kind: Deployment
metadata:
  name: x
`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported kind")
}

func TestParseDataFlow_InvalidYAML(t *testing.T) {
	_, err := ParseDataFlow([]byte("invalid: yaml: ["))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parse manifest")
}
