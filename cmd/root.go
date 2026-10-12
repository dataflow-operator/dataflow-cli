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
	"os"

	"github.com/dataflow-operator/dataflow-cli/internal/apiclient"
	"github.com/spf13/cobra"
)

var (
	apiURL     string
	namespace  string
	defaultNS  = "default"
	apiURLEnv  = "DATAFLOW_API_URL"
	defaultAPI = "http://localhost:8080/api"
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "dataflow-cli",
	Short: "CLI for DataFlow operator: Web API client and local manifest run",
	Long:  `dataflow-cli talks to the dataflow-web API (list/create/update/delete dataflows, logs, status, metrics) and can run DataFlow manifests locally without Kubernetes.`,
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&apiURL, "api-url", defaultAPI, "Base URL of the dataflow-web API (env: "+apiURLEnv+")")
	rootCmd.PersistentFlags().StringVar(&namespace, "namespace", defaultNS, "Kubernetes namespace for dataflow resources")

	rootCmd.AddCommand(namespacesCmd)
	rootCmd.AddCommand(dataflowsCmd)
	rootCmd.AddCommand(logsCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(metricsCmd)
}

func getClient() *apiclient.Client {
	return apiclient.New(getAPIURL())
}

func getAPIURL() string {
	if e := os.Getenv(apiURLEnv); e != "" {
		return e
	}
	return apiURL
}

func getNamespace() string {
	if namespace != "" {
		return namespace
	}
	return defaultNS
}
