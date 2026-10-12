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
	"encoding/json"
	"fmt"
	"os"

	dataflowv1 "github.com/dataflow-operator/dataflow/api/v1"
	"sigs.k8s.io/yaml"
)

const (
	KindDataFlow     = "DataFlow"
	KindDataFlowCron = "DataFlowCron"
)

// Pipeline is a local-run view of a DataFlow or DataFlowCron manifest.
// Cron-only fields (schedule, triggers) are not kept.
type Pipeline struct {
	APIVersion string
	Kind       string
	Name       string
	Namespace  string
	Spec       dataflowv1.DataFlowSpec
}

// LoadDataFlow reads a YAML file and parses it as a DataFlow custom resource.
func LoadDataFlow(path string) (*dataflowv1.DataFlow, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read manifest: %w", err)
	}
	return ParseDataFlow(data)
}

// ParseDataFlow parses YAML bytes as a DataFlow custom resource.
func ParseDataFlow(data []byte) (*dataflowv1.DataFlow, error) {
	var df dataflowv1.DataFlow
	if err := yaml.Unmarshal(data, &df); err != nil {
		return nil, fmt.Errorf("parse manifest: %w", err)
	}
	return &df, nil
}

// LoadPipeline reads a YAML file as a DataFlow or DataFlowCron manifest.
func LoadPipeline(path string) (*Pipeline, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read manifest: %w", err)
	}
	return ParsePipeline(data)
}

// ParsePipeline parses YAML bytes as a DataFlow or DataFlowCron manifest.
func ParsePipeline(data []byte) (*Pipeline, error) {
	var raw struct {
		APIVersion string `json:"apiVersion"`
		Kind       string `json:"kind"`
		Metadata   struct {
			Name      string `json:"name"`
			Namespace string `json:"namespace"`
		} `json:"metadata"`
		Spec json.RawMessage `json:"spec"`
	}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parse manifest: %w", err)
	}
	if raw.APIVersion != "" && raw.APIVersion != dataflowv1.GroupVersion.String() {
		return nil, fmt.Errorf("unsupported apiVersion %q", raw.APIVersion)
	}
	switch raw.Kind {
	case KindDataFlow, KindDataFlowCron:
	default:
		return nil, fmt.Errorf("unsupported kind %q: want DataFlow or DataFlowCron", raw.Kind)
	}
	if raw.Metadata.Name == "" {
		return nil, fmt.Errorf("metadata.name is required")
	}
	var spec dataflowv1.DataFlowSpec
	if len(raw.Spec) > 0 {
		if err := yaml.Unmarshal(raw.Spec, &spec); err != nil {
			return nil, fmt.Errorf("parse spec: %w", err)
		}
	}
	return &Pipeline{
		APIVersion: raw.APIVersion,
		Kind:       raw.Kind,
		Name:       raw.Metadata.Name,
		Namespace:  raw.Metadata.Namespace,
		Spec:       spec,
	}, nil
}
