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

package gengo

import (
	"bytes"
	"fmt"
	ttmarkers "github.com/Raskyld/go-tektasker/pkg/markers"
	"go/ast"
	"log/slog"
	"sigs.k8s.io/controller-tools/pkg/genall"
	"sigs.k8s.io/controller-tools/pkg/markers"
	"text/template"
)

const FuncName = "func"
const FuncTpl = `{{template "%s" .Header}}

{{- range $tplName, $args := .Templates}}
{{- range $args}}
{{CallTemplate $tplName .}}
{{- end}}
{{- end}}
`

type TaskGoFuncGenerator struct {
	Logger *slog.Logger

	Template *template.Template
}

// PerTemplateArgs maps each registered template to a mapping of param or result name to template args
// templateName -> Param/Result Name -> Param/ResultArgs
// We then will execute each template for each registered param or name with
// the associated args
type PerTemplateArgs map[string]map[string]interface{}

type FuncArgs struct {
	Header    GoHeaderArgs
	Templates PerTemplateArgs
}

func NewGoFunc(logger *slog.Logger) (*TaskGoFuncGenerator, error) {
	g := &TaskGoFuncGenerator{
		Logger:   logger.With("generator", "goFunc"),
		Template: &template.Template{},
	}

	// NB(raskyld): this hack is needed to call template
	// with name resolved at runtime as template/text does check
	// template references at parsing time
	g.Template.Funcs(template.FuncMap{
		"CallTemplate": func(tplName string, args interface{}) string {
			var buf bytes.Buffer
			err := g.Template.ExecuteTemplate(&buf, tplName, args)
			if err != nil {
				panic(err)
			}
			return buf.String()
		},
	})

	g.RegisterTemplate(GoHeaderName, GoHeaderTpl).
		RegisterTemplate(ParamFuncNameName, ParamFuncNameTpl).
		RegisterTemplate(ParamFuncUnmarshalSimpleName, ParamFuncUnmarshalSimpleTpl).
		RegisterTemplate(ParamFuncUnmarshalJSONName, ParamFuncUnmarshalJSONTpl).
		RegisterTemplate(ResultFuncNameName, ResultFuncNameTpl).
		RegisterTemplate(ResultFuncMarshalSimpleName, ResultFuncMarshalSimpleTpl).
		RegisterTemplate(ResultFuncMarshalJSONName, ResultFuncMarshalJSONTpl).
		RegisterTemplate(FuncName, fmt.Sprintf(FuncTpl, GoHeaderName))

	return g, nil
}

func (g *TaskGoFuncGenerator) RegisterTemplate(name string, tpl string) *TaskGoFuncGenerator {
	nt, err := g.Template.New(name).Parse(tpl)
	if err != nil {
		panic(err)
	}

	g.Template = nt
	return g
}

func (*TaskGoFuncGenerator) RegisterMarkers(into *markers.Registry) error {
	return ttmarkers.Register(into)
}

func (g *TaskGoFuncGenerator) Generate(ctx *genall.GenerationContext) error {
	for _, pkg := range ctx.Roots {
		logger := g.Logger.With("pkg", pkg.Name)
		logger.Debug("starting collecting")

		// Check the package-level task marker is present or skip
		pkgMarkers, err := markers.PackageMarkers(ctx.Collector, pkg)
		if err != nil {
			return err
		}

		taskMarker := pkgMarkers.Get(ttmarkers.MarkerTask)
		if taskMarker == nil {
			// If no task marker is set on package, simply skip it
			logger.Info("skipping non-task package")
			continue
		}

		perTemplateArgs := PerTemplateArgs(make(map[string]map[string]interface{}))
		for _, t := range g.Template.Templates() {
			// NB(raskyld): we need to exclude the main template as
			// this would lead to infinite recursive call as we iterate
			// over the map perTemplateArgs and call every template
			if t.Name() == FuncName {
				continue
			}

			perTemplateArgs[t.Name()] = make(map[string]interface{})
		}

		err = markers.EachType(ctx.Collector, pkg, func(info *markers.TypeInfo) {
			rawParam := info.Markers.Get(ttmarkers.MarkerParam)
			if rawParam != nil {
				if param, ok := rawParam.(ttmarkers.Param); ok {
					logger := logger.With("param", param.Name)
					logger.Info("parameter found")

					// ensure no duplication
					if _, duplicate := perTemplateArgs[ParamFuncNameName][param.Name]; duplicate {
						logger.Warn("parameter duplicated! ensure unique name")
						return
					}

					perTemplateArgs[ParamFuncNameName][param.Name] = ParamFuncArgs{
						ParamName: param.Name,
						ParamType: info.Name,
					}

					// TODO(raskyld):
					// 	add a way to skip creating the Unmarshal method so our
					// 	users can create a custom unmarshaler for their complex types

					mapToAdd := ParamFuncUnmarshalJSONName

					// NB(raskyld): this is for ease of use when we have a type made of string
					// otherwise, we just expect the value to be valid JSON
					if ident, ok := info.RawSpec.Type.(*ast.Ident); ok {
						if ident.Name == "string" {
							mapToAdd = ParamFuncUnmarshalSimpleName
						}
					}

					perTemplateArgs[mapToAdd][param.Name] = ParamFuncArgs{
						ParamName: param.Name,
						ParamType: info.Name,
					}
				}
			}

			rawResult := info.Markers.Get(ttmarkers.MarkerResult)
			if rawResult != nil {
				if result, ok := rawResult.(ttmarkers.Result); ok {
					logger := logger.With("result", result.Name)
					logger.Info("result found")

					// ensure no duplication
					if _, duplicate := perTemplateArgs[ResultFuncNameName][result.Name]; duplicate {
						logger.Warn("result duplicated! ensure unique name")
						return
					}

					perTemplateArgs[ResultFuncNameName][result.Name] = ResultFuncArgs{
						ResultName: result.Name,
						ResultType: info.Name,
					}

					mapToAdd := ResultFuncMarshalJSONName

					// NB(raskyld): this is for ease of use when we have a type made of string
					// otherwise, we just expect the value to be valid JSON
					if ident, ok := info.RawSpec.Type.(*ast.Ident); ok {
						if ident.Name == "string" {
							mapToAdd = ResultFuncMarshalSimpleName
						}
					}

					perTemplateArgs[mapToAdd][result.Name] = ResultFuncArgs{
						ResultName: result.Name,
						ResultType: info.Name,
					}
				}
			}
		})

		output, err := ctx.OutputRule.Open(pkg, "zz_generated.tektasker.go")
		if err != nil {
			return err
		}

		importPaths := make([]string, 0)
		if len(perTemplateArgs[ResultFuncMarshalJSONName]) > 0 || len(perTemplateArgs[ParamFuncUnmarshalJSONName]) > 0 {
			importPaths = append(importPaths, "encoding/json")
		}

		err = g.Template.ExecuteTemplate(output, FuncName, FuncArgs{
			Header: GoHeaderArgs{
				PkgName:     pkg.Name,
				Header:      "// generated code",
				ImportPaths: importPaths,
			},
			Templates: perTemplateArgs,
		})

		if err != nil {
			return err
		}
	}

	return nil
}
