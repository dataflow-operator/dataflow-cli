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
	"encoding/json"
	"fmt"
	"os"

	"github.com/dataflow-operator/dataflow-cli/internal/manifest"
	"github.com/spf13/cobra"
)

var (
	dataflowFile string
)

var dataflowsCmd = &cobra.Command{
	Use:   "dataflows",
	Short: "Manage DataFlows via API (list, get, create, update, delete)",
}

var dataflowsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List DataFlows in a namespace",
	RunE:  runDataflowsList,
}

var dataflowsGetCmd = &cobra.Command{
	Use:   "get [name]",
	Short: "Get a DataFlow by name",
	Args:  cobra.ExactArgs(1),
	RunE:  runDataflowsGet,
}

var dataflowsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a DataFlow from a manifest file",
	RunE:  runDataflowsCreate,
}

var dataflowsUpdateCmd = &cobra.Command{
	Use:   "update [name]",
	Short: "Update a DataFlow from a manifest file",
	Args:  cobra.ExactArgs(1),
	RunE:  runDataflowsUpdate,
}

var dataflowsDeleteCmd = &cobra.Command{
	Use:   "delete [name]",
	Short: "Delete a DataFlow",
	Args:  cobra.ExactArgs(1),
	RunE:  runDataflowsDelete,
}

func init() {
	dataflowsCreateCmd.Flags().StringVarP(&dataflowFile, "file", "f", "", "Path to DataFlow manifest YAML (required)")
	_ = dataflowsCreateCmd.MarkFlagRequired("file")
	dataflowsUpdateCmd.Flags().StringVarP(&dataflowFile, "file", "f", "", "Path to DataFlow manifest YAML (required)")
	_ = dataflowsUpdateCmd.MarkFlagRequired("file")

	dataflowsCmd.AddCommand(dataflowsListCmd, dataflowsGetCmd, dataflowsCreateCmd, dataflowsUpdateCmd, dataflowsDeleteCmd)
}

func runDataflowsList(cmd *cobra.Command, args []string) error {
	client := getClient()
	list, err := client.ListDataFlows(getNamespace())
	if err != nil {
		return err
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(list)
}

func runDataflowsGet(cmd *cobra.Command, args []string) error {
	client := getClient()
	name := args[0]
	out, err := client.GetDataFlow(getNamespace(), name)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}

func runDataflowsCreate(cmd *cobra.Command, args []string) error {
	df, err := manifest.LoadDataFlow(dataflowFile)
	if err != nil {
		return err
	}
	df.Namespace = getNamespace()
	body, err := json.Marshal(df)
	if err != nil {
		return err
	}
	client := getClient()
	out, err := client.CreateDataFlow(getNamespace(), body)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}

func runDataflowsUpdate(cmd *cobra.Command, args []string) error {
	name := args[0]
	df, err := manifest.LoadDataFlow(dataflowFile)
	if err != nil {
		return err
	}
	df.Name = name
	df.Namespace = getNamespace()
	body, err := json.Marshal(df)
	if err != nil {
		return err
	}
	client := getClient()
	out, err := client.UpdateDataFlow(getNamespace(), name, body)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}

func runDataflowsDelete(cmd *cobra.Command, args []string) error {
	client := getClient()
	if err := client.DeleteDataFlow(getNamespace(), args[0]); err != nil {
		return err
	}
	fmt.Println("deleted")
	return nil
}
