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
    initGenerator "github.com/Raskyld/go-tektasker/internal/init"
    "github.com/spf13/cobra"
    "sigs.k8s.io/controller-tools/pkg/genall"
)

func NewInit(ctx *Context) *cobra.Command {
    cmd := &cobra.Command{
        Use:   "init task-name [output]",
        Short: "Init an opinionated project to write a Task in Go",
        Args:  cobra.RangeArgs(1, 2),
        RunE: func(cmd *cobra.Command, args []string) error {
            taskName := "default"

            if len(args) > 0 {
                taskName = args[0]
            }

            var outRule genall.OutputRule
            if ctx.DryRun {
                outRule = genall.OutputToStdout
            } else if len(args) > 1 {
                outRule = genall.OutputToDirectory(args[1])
            } else {
                outRule = genall.OutputToDirectory(".")
            }

            gen := initGenerator.New(ctx.Logger, taskName)

            err := gen.Generate(outRule)
            if err != nil {
                return err
            }

            ctx.Logger.Info("project init has been successful, please check the .env file for further configuration")

            return nil
        },
    }

    return cmd
}
