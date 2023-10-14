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
	"github.com/Raskyld/go-tektasker/internal/gengo"
	"github.com/Raskyld/go-tektasker/internal/genyaml"
	"github.com/spf13/cobra"
	"path/filepath"
	"sigs.k8s.io/controller-tools/pkg/genall"
)

func NewGenerate(ctx *Context) *cobra.Command {
	generate := &cobra.Command{
		Use:     "generate",
		Aliases: []string{"gen", "g"},
		Short:   "Generate artifacts from the specified package",
	}

	BindGenerate(ctx, generate)

	generate.AddCommand(NewGenerateManifest(ctx))
	generate.AddCommand(NewGenerateGo(ctx))

	return generate
}

func NewGenerateGo(ctx *Context) *cobra.Command {
	genFuncGo := &cobra.Command{
		Use:   "go [internalPkgPath] [internalPkgName]",
		Short: "Generate the Go code to integrate with Tekton",
		Long: `This command when run without args is equivalent to

tektasker gen go internal/ tekton

Which means, the helper functions are gonna be written under internal/tekton
directory with a package name of "tekton".

The code generated for your main package will be written in zz_generated.tektasker.go
`,
		Args: cobra.MaximumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			outputInternal := "internal/"
			outputPkgName := "tekton"

			if len(args) > 0 {
				outputInternal = args[0]
			}

			if len(args) > 1 {
				outputPkgName = args[1]
			}

			var genFunc, genInternal genall.Generator

			genFuncPtr, err := gengo.NewGoFunc(ctx.Logger)
			if err != nil {
				return err
			}

			genFunc = genFuncPtr

			genInternalPtr, err := gengo.NewGoInternal(ctx.Logger, outputPkgName)
			if err != nil {
				return err
			}

			genInternal = genInternalPtr

			gens := genall.Generators{&genFunc, &genInternal}

			runtime, err := gens.ForRoots(ctx.Generate.Input)
			if err != nil {
				return err
			}

			if ctx.DryRun {
				runtime.OutputRules = genall.OutputRules{
					Default:     genall.OutputToStdout,
					ByGenerator: nil,
				}
			} else {
				runtime.OutputRules = genall.OutputRules{
					Default: genall.OutputArtifacts{},
					ByGenerator: map[*genall.Generator]genall.OutputRule{
						&genInternal: genall.OutputToDirectory(filepath.Join(outputInternal, outputPkgName)),
					},
				}
			}

			runtime.Run()
			return nil
		},
	}

	return genFuncGo
}

func NewGenerateManifest(ctx *Context) *cobra.Command {
	genYaml := &cobra.Command{
		Use:   "manifest output-dir",
		Short: "Generate your YAML manifests and write them in the given output-dir",
		Example: `
# Generate Task manifest for the current working directory package
tektasker gen manifest ./manifests/

# Generate Task manifest for a specific package
tektasker gen -i ./pkg/helloworld manifest ./manifests/

# Generate Task for every package in pkg/
tektasker gen -i ./pkg/... manifest ./manifests/
`,
		Args: ResolveManifestArgs(ctx),
		RunE: func(cmd *cobra.Command, args []string) error {
			var gen genall.Generator = genyaml.TaskYamlGenerator{ctx.Logger}
			gens := genall.Generators{&gen}

			runtime, err := gens.ForRoots(ctx.Generate.Input)
			if err != nil {
				return err
			}

			var outRule genall.OutputRule
			if ctx.DryRun {
				outRule = genall.OutputToStdout
			} else {
				outRule = genall.OutputToDirectory(args[0])
			}

			runtime.OutputRules = genall.OutputRules{
				Default: outRule,
			}

			runtime.Run()
			return nil
		},
	}

	return genYaml
}

// ResolveManifestArgs resolves how many argument is accepted by the command at runtime
func ResolveManifestArgs(ctx *Context) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if ctx.DryRun {
			return cobra.MaximumNArgs(1)(cmd, args)
		} else {
			return cobra.ExactArgs(1)(cmd, args)
		}
	}
}
