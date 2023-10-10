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

package genyaml

import (
	"errors"
	ttmarkers "github.com/Raskyld/go-tektasker/pkg/markers"
	v1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	"go/ast"
	"log/slog"
	"sigs.k8s.io/controller-tools/pkg/genall"
	"sigs.k8s.io/controller-tools/pkg/markers"
)

const KubernetesVersionLabel = "app.kubernetes.io/version"

type TaskYamlGenerator struct {
	Logger *slog.Logger
}

func (TaskYamlGenerator) RegisterMarkers(into *markers.Registry) error {
	return ttmarkers.Register(into)
}

func (g TaskYamlGenerator) Generate(ctx *genall.GenerationContext) error {
	for _, pkg := range ctx.Roots {
		// TODO(raskyld): get all package level markers
		// NB(raskyld): in the future we may/should use an IR to avoid hard coupling
		// between the generation process and the specific v1 version
		logger := g.Logger.With("pkg", pkg.Name)

		logger.Debug("starting collecting")

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

		var task v1.Task

		if taskMarker, ok := taskMarker.(ttmarkers.Task); ok {
			task.SetName(taskMarker.Name)

			if task.Labels == nil {
				task.Labels = make(map[string]string)
			}

			task.Labels[KubernetesVersionLabel] = taskMarker.Version
		} else {
			return errors.New("unexpected wrong type for task marker")
		}

		// we keep a mapping from param and result name to scheme index
		// to avoid duplicates
		paramsIdx := make(map[string]int)

		params := &task.Spec.Params

		err = markers.EachType(ctx.Collector, pkg, func(info *markers.TypeInfo) {
			rawParam := info.Markers.Get(ttmarkers.MarkerParam)
			if rawParam != nil {
				if param, ok := rawParam.(ttmarkers.Param); ok {
					logger := logger.With("param", param.Name)
					logger.Info("parameter found")

					// ensure no duplication
					if _, duplicate := paramsIdx[param.Name]; duplicate {
						logger.Warn("parameter duplicated! ensure unique name")
						return
					}

					paramsIdx[param.Name] = len(*params)
					builtParam, err := g.buildParam(param, info)
					if err != nil {
						logger.Warn("could not create param", "err", err)
						return
					}

					*params = append(*params, builtParam)
				}
			}
		})

		if err != nil {
			return err
		}

		// TODO(raskyld): Then for each params and results create the needed ENV var
		err = ctx.WriteYAML(task.Name+"-task.yaml", "", []interface{}{task})
		if err != nil {
			return err
		}
	}

	return nil
}

func (g TaskYamlGenerator) buildParam(param ttmarkers.Param, typeInfo *markers.TypeInfo) (v1.ParamSpec, error) {
	// First, we must figure out which Tekton type to use for the marked type
	var tektonType v1.ParamType
	switch typeInfo.RawSpec.Type.(type) {
	case *ast.ArrayType:
		tektonType = v1.ParamTypeArray
	case *ast.MapType, *ast.StructType:
		tektonType = v1.ParamTypeObject
	default:
		tektonType = v1.ParamTypeString
	}

	rt := v1.ParamSpec{
		Name:        param.Name,
		Type:        tektonType,
		Description: typeInfo.Doc,
	}

	if param.Default != nil {
		defaultValue := &v1.ParamValue{}
		err := defaultValue.UnmarshalJSON([]byte(*param.Default))
		if err != nil {
			return rt, err
		}

		if defaultValue.Type != tektonType {
			return rt, errors.New("type mismatch between default value and Go type")
		}

		rt.Default = defaultValue
	}

	return rt, nil
}
