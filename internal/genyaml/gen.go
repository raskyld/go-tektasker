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
	"encoding/json"
	"errors"
	ttmarkers "github.com/Raskyld/go-tektasker/pkg/markers"
	"go/ast"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"log/slog"
	"sigs.k8s.io/controller-tools/pkg/genall"
	"sigs.k8s.io/controller-tools/pkg/markers"
	"strings"
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
		// NB(raskyld): in the future we may/should use an IR to avoid hard coupling
		// between the generation process and the specific v1 version
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

		var task unstructured.Unstructured
		if taskMarker, ok := taskMarker.(ttmarkers.Task); ok {
			task.SetName(taskMarker.Name)
			task.SetLabels(map[string]string{
				KubernetesVersionLabel: taskMarker.Version,
			})
		} else {
			return errors.New("unexpected wrong type for task marker")
		}

		// For now we only support the v1 api
		task.SetGroupVersionKind(schema.GroupVersionKind{
			Group:   "tekton.dev",
			Version: "v1",
			Kind:    "Task",
		})

		// Concatenate every package-level doc from all files
		// to create the description of the task
		packagesDoc := make([]string, 0, len(pkg.Syntax))
		for _, file := range pkg.Syntax {
			if file != nil && file.Doc != nil {
				packagesDoc = append(packagesDoc, file.Doc.Text())
			}
		}

		if len(packagesDoc) != 0 {
			err = unstructured.SetNestedField(task.Object, strings.Join(packagesDoc, "\n"), "spec", "description")
			if err != nil {
				return err
			}
		}

		// Generate workspaces
		if workspaces, ok := pkgMarkers[ttmarkers.MarkerWorkspace]; ok {
			workspacesYaml := make([]interface{}, 0, len(workspaces))
			for _, workspace := range workspaces {
				if workspace, isWorkspace := workspace.(ttmarkers.Workspace); isWorkspace {
					logger.Info("found workspace", "workspace", workspace.Name)
					workspaceYaml, err := g.buildWorkspace(workspace)

					if err != nil {
						return err
					}

					workspacesYaml = append(workspacesYaml, workspaceYaml)
				}
			}

			err := unstructured.SetNestedSlice(task.Object, workspacesYaml, "spec", "workspaces")
			if err != nil {
				return err
			}
		}

		// we keep a mapping from param and result name to scheme index
		// to avoid duplicates
		paramsIdx := make(map[string]int)
		params := make([]interface{}, 0)
		resultsIdx := make(map[string]int)
		results := make([]interface{}, 0)

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

					paramsIdx[param.Name] = len(params)
					builtParam, err := g.buildParam(param, info)
					if err != nil {
						logger.Warn("could not create param", "err", err)
						return
					}

					params = append(params, builtParam)
				}
			}

			rawResult := info.Markers.Get(ttmarkers.MarkerResult)
			if rawResult != nil {
				if result, ok := rawResult.(ttmarkers.Result); ok {
					logger := logger.With("result", result.Name)
					logger.Info("result found")

					// ensure no duplication
					if _, duplicate := resultsIdx[result.Name]; duplicate {
						logger.Warn("result duplicated! ensure unique name")
						return
					}

					resultsIdx[result.Name] = len(results)
					builtResult, err := g.buildResult(result, info)
					if err != nil {
						logger.Warn("could not create result", "err", err)
						return
					}

					results = append(results, builtResult)
				}
			}
		})

		if err != nil {
			return err
		}

		err = unstructured.SetNestedSlice(task.Object, params, "spec", "params")
		if err != nil {
			return err
		}

		err = unstructured.SetNestedSlice(task.Object, results, "spec", "results")
		if err != nil {
			return err
		}

		// TODO(raskyld): Then for each params and results create the needed ENV var
		err = ctx.WriteYAML(task.GetName()+"-task.yaml", "", []interface{}{task.Object})
		if err != nil {
			return err
		}
	}

	return nil
}

func (g TaskYamlGenerator) buildParam(param ttmarkers.Param, typeInfo *markers.TypeInfo) (map[string]interface{}, error) {
	rt := map[string]interface{}{
		"name":        param.Name,
		"description": typeInfo.Doc,
	}

	// First, we must figure out which Tekton type to use for the marked type
	var tektonType string
	switch typeInfo.RawSpec.Type.(type) {
	case *ast.ArrayType:
		tektonType = "array"
	case *ast.MapType:
		// Should be a valid json string though as
		// we don't have a non-strict json schema type in Tekton
		tektonType = "string"
	case *ast.StructType:
		// If the struct is not in strict mode then it does not have
		// a determinist schema
		if !param.Strict {
			tektonType = "string"
			break
		}

		properties := make(map[string]interface{})
		for _, field := range typeInfo.Fields {
			tag, hasTag := field.Tag.Lookup("json")
			if !hasTag {
				return nil, errors.New("missing json tag on your strict struct")
			}

			tags := strings.Split(tag, ",")
			if len(tags[0]) == 0 {
				return nil, errors.New("you need to give explicit json tag name to your struct")
			} else {
				if strings.HasPrefix(tags[0], "-") {
					// Ignore this field
					continue
				} else {
					if _, ok := properties[tags[0]]; ok {
						return nil, errors.New("cannot use duplicate json tag name")
					}

					properties[tags[0]] = map[string]interface{}{
						"type": "string",
					}
				}
			}
		}

		rt["properties"] = properties

		tektonType = "object"
	default:
		tektonType = "string"
	}

	rt["type"] = tektonType

	if param.Default != nil {
		defValue := *param.Default
		switch tektonType {
		case "array":
			var defaultArray []interface{}
			err := json.Unmarshal([]byte(defValue), &defaultArray)
			rt["default"] = defaultArray
			return rt, err
		case "object":
			// See TEP-0075 for how we should update and maintain this section
			var object map[string]interface{}
			err := json.Unmarshal([]byte(defValue), &object)
			rt["default"] = object
			return rt, err
		default:
			rt["default"] = defValue
		}
	}

	return rt, nil
}

func (g TaskYamlGenerator) buildResult(result ttmarkers.Result, typeInfo *markers.TypeInfo) (map[string]interface{}, error) {
	rt := map[string]interface{}{
		"name":        result.Name,
		"description": typeInfo.Doc,
	}

	// First, we must figure out which Tekton type to use for the marked type
	var tektonType string
	switch typeInfo.RawSpec.Type.(type) {
	case *ast.ArrayType:
		tektonType = "array"
	default:
		tektonType = "string"
	}

	rt["type"] = tektonType

	return rt, nil
}

func (g TaskYamlGenerator) buildWorkspace(workspace ttmarkers.Workspace) (map[string]interface{}, error) {
	rt := map[string]interface{}{
		"name":        workspace.Name,
		"description": workspace.Description,
		"readOnly":    workspace.ReadOnly,
		"optional":    workspace.Optional,
	}

	if workspace.MountPath != "" {
		rt["mountPath"] = workspace.MountPath
	}

	return rt, nil
}
