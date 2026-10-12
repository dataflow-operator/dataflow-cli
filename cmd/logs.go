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
	"bufio"
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

var (
	logsTail  int
	logsFollow bool
)

var logsCmd = &cobra.Command{
	Use:   "logs [name]",
	Short: "Get logs for a DataFlow processor",
	Args:  cobra.ExactArgs(1),
	RunE:  runLogs,
}

func init() {
	logsCmd.Flags().IntVar(&logsTail, "tail", 100, "Number of lines to show (non-follow)")
	logsCmd.Flags().BoolVarP(&logsFollow, "follow", "f", false, "Stream logs (SSE)")
}

func runLogs(cmd *cobra.Command, args []string) error {
	name := args[0]
	client := getClient()
	if logsFollow {
		stream, err := client.OpenLogStream(getNamespace(), name, logsTail)
		if err != nil {
			return err
		}
		defer stream.Close()
		return streamLogs(stream)
	}
	text, err := client.GetLogs(getNamespace(), name, logsTail)
	if err != nil {
		return err
	}
	fmt.Print(text)
	return nil
}

// streamLogs reads SSE-style "data: <line>\n\n" and prints each log line.
func streamLogs(r io.Reader) error {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) >= 6 && line[:6] == "data: " {
			fmt.Println(line[6:])
		}
	}
	return scanner.Err()
}
