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
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/dataflow-operator/dataflow-cli/internal/manifest"
	runpkg "github.com/dataflow-operator/dataflow/pkg/run"
	"github.com/spf13/cobra"
)

var (
	summaryOut io.Writer = os.Stdout
	warnOut    io.Writer = os.Stderr
)

var (
	runLogLevel       string
	runOnce           bool
	runCheckpointFile string
	runReset          bool
	runTimeout        time.Duration
)

// executePipeline runs a spec in-process. Tests replace it.
var executePipeline = runpkg.Run

var runCmd = &cobra.Command{
	Use:   "run [manifest.yaml]",
	Short: "Run a DataFlow manifest locally (no Kubernetes)",
	Long: `Loads a DataFlow or DataFlowCron manifest and runs the pipeline in-process.
Source and sink must be reachable (for example Kafka or Postgres). Stop with Ctrl+C.

--once runs until a polling source is exhausted (postgresql, clickhouse, trino, iceberg, nessie, hive),
writes a local checkpoint file, and exits. Kafka and postgresql-cdc are rejected.
DataFlowCron schedule and triggers are not executed.`,
	Args: cobra.ExactArgs(1),
	RunE: runRun,
}

func init() {
	runCmd.Flags().StringVar(&runLogLevel, "log-level", "info", "Log level: debug, info, warn, error")
	runCmd.Flags().BoolVar(&runOnce, "once", false, "Run until the source is exhausted, then exit")
	runCmd.Flags().StringVar(&runCheckpointFile, "checkpoint-file", "", "Local checkpoint JSON file (default for --once: .dataflow/<name>.json)")
	runCmd.Flags().BoolVar(&runReset, "reset", false, "Clear the source checkpoint before start")
	runCmd.Flags().DurationVar(&runTimeout, "timeout", 0, "Maximum run time (for example 2h); 0 means no limit")
	rootCmd.AddCommand(runCmd)
}

func runRun(cmd *cobra.Command, args []string) error {
	pipeline, err := manifest.LoadPipeline(args[0])
	if err != nil {
		return err
	}
	if pipeline.Kind == manifest.KindDataFlowCron {
		fmt.Fprintln(warnOut, "warning: DataFlowCron schedule and triggers are ignored; the pipeline runs in-process and triggers are not executed")
	}

	ns := pipeline.Namespace
	if ns == "" {
		ns = getNamespace()
	}
	checkpointPath := runCheckpointFile
	if runOnce && checkpointPath == "" {
		checkpointPath = defaultCheckpointPath(pipeline.Name)
	}
	if runReset && checkpointPath == "" {
		return fmt.Errorf("--reset requires --once or --checkpoint-file")
	}

	ctx := context.Background()
	var cancel context.CancelFunc
	if runTimeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, runTimeout)
	} else {
		ctx, cancel = context.WithCancel(ctx)
	}
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigCh)
	go func() {
		select {
		case <-sigCh:
			cancel()
		case <-ctx.Done():
		}
	}()

	opts := runpkg.RunOptions{
		Namespace:       ns,
		Name:            pipeline.Name,
		LogLevel:        runLogLevel,
		CheckpointPath:  checkpointPath,
		ResetCheckpoint: runReset,
		Once:            runOnce,
	}
	result, err := executePipeline(ctx, &pipeline.Spec, opts)
	if runOnce && (err == nil || result.CheckpointPath != "" || result.Processed != 0 || result.Errors != 0) {
		printRunSummary(summaryOut, pipeline, result)
	}
	return err
}

func defaultCheckpointPath(name string) string {
	return filepath.Join(".dataflow", name+".json")
}

func printRunSummary(w io.Writer, pipeline *manifest.Pipeline, result runpkg.RunResult) {
	checkpoint := result.CheckpointPath
	if checkpoint == "" {
		checkpoint = "(disabled)"
	}
	fmt.Fprintf(w, "source: %s\nsink: %s\nprocessed: %d\nerrors: %d\ncheckpoint: %s\n",
		pipeline.Spec.Source.Type,
		pipeline.Spec.Sink.Type,
		result.Processed,
		result.Errors,
		checkpoint,
	)
}
