/*
Copyright 2023 Enzo Nocera <enzo@nocera.eu>

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
	"github.com/spf13/cobra"
	"log/slog"
	"os"
)

func New() *cobra.Command {
	var debug bool
	var ctx Context

	root := &cobra.Command{
		Use:   "tektasker",
		Short: "From your Go IDE to your Tekton Cluster in minutes",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			handlerOpt := &slog.HandlerOptions{}
			if debug {
				handlerOpt.Level = slog.LevelDebug
			}

			ctx.Logger = slog.New(slog.NewTextHandler(os.Stderr, handlerOpt))
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	root.PersistentFlags().BoolVar(&debug, "debug", false, "run in debug mode")
	root.PersistentFlags().BoolVar(&ctx.DryRun, "dry-run", false, "only output to stdout")

	root.AddCommand(NewGenerate(&ctx))
	root.AddCommand(NewMarkers(&ctx))

	return root
}
