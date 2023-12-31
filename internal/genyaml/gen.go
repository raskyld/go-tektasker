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
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	ttmarkers "github.com/raskyld/go-tektasker/pkg/markers"
	"go/ast"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"log/slog"
	"path"
	"path/filepath"
	"sigs.k8s.io/controller-tools/pkg/genall"
	"sigs.k8s.io/controller-tools/pkg/loader"
	"sigs.k8s.io/controller-tools/pkg/markers"
	"strconv"
	"strings"
	"text/template"
)

const KubernetesVersionLabel = "app.kubernetes.io/version"

type TaskYamlGenerator struct {
	Logger *slog.Logger

	// StepCommand is used to fix the command field of the step (i.e. the entrypoint of your container)
	StepCommand string
}

// stepCommandArgs is passed when templating the value of StepCommand
type stepCommandArgs struct {
	// PkgName is the name of the package being parsed
	PkgName string

	// ImportPath of the package
	ImportPath string

	// TODO(raskyld): Fragile code! We should use a custom type capable of replicating the behavior of various ko version
	// KoAppName is replicating the way ko builder (0.15) determinate the entrypoint of the built container
	// See issue #15 for details
	KoAppName string
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

		task, err := g.initTask(taskMarker)
		if err != nil {
			return err
		}

		err = g.buildPackageDoc(task, pkg)
		if err != nil {
			return err
		}

		err = g.buildWorkspaces(task, pkgMarkers)
		if err != nil {
			return err
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

		err = g.buildSteps(task, pkg, params, results)
		if err != nil {
			return err
		}

		kustomization := map[string]interface{}{
			"resources": []interface{}{
				task.GetName() + "-task.yaml",
			},
		}

		err = ctx.WriteYAML("base/"+task.GetName()+"-task.yaml", "", []interface{}{task.Object})
		if err != nil {
			return err
		}

		err = ctx.WriteYAML("base/kustomization.yaml", "", []interface{}{kustomization})
		if err != nil {
			return err
		}
	}

	return nil
}

func (g TaskYamlGenerator) buildSteps(task unstructured.Unstructured, pkg *loader.Package, params, results []interface{}) error {
	mainStep := map[string]interface{}{
		"image": "ko://" + pkg.PkgPath,
	}

	// NB(raskyld): this code is dirty asf, we should be able to clean it when
	// migrating from ad-hoc generation to intermediate representation (IR)

	command, err := g.parseFlagCommand(pkg)
	if err != nil {
		return err
	}

	cmds := strings.Split(command, " ")
	icmds := make([]interface{}, len(cmds))
	for i, cmd := range cmds {
		icmds[i] = cmd
	}

	mainStep["command"] = icmds

	envs := make([]interface{}, 0)
	for _, param := range params {
		if param, ok := param.(map[string]interface{}); ok {
			paramName := param["name"].(string)
			envVarName := fmt.Sprintf("PARAM_%s_VALUE", strings.ToUpper(paramName))
			paramValue := fmt.Sprintf("$(params[%s]", strconv.Quote(paramName))

			if param["type"].(string) != "string" {
				paramValue += "[*])"
			} else {
				paramValue += ")"
			}

			envs = append(envs, map[string]interface{}{
				"name":  envVarName,
				"value": paramValue,
			})
		}
	}

	for _, result := range results {
		if result, ok := result.(map[string]interface{}); ok {
			resultName := result["name"].(string)
			envVarName := fmt.Sprintf("RESULT_%s_PATH", strings.ToUpper(resultName))
			resultValue := fmt.Sprintf("$(results[%s].path)", strconv.Quote(resultName))

			envs = append(envs, map[string]interface{}{
				"name":  envVarName,
				"value": resultValue,
			})
		}
	}

	if len(envs) > 0 {
		mainStep["env"] = envs
	}

	return unstructured.SetNestedSlice(task.Object, []interface{}{mainStep}, "spec", "steps")
}

// parseFlagCommand parses the flag `command` passed by the user
func (g TaskYamlGenerator) parseFlagCommand(pkg *loader.Package) (string, error) {
	tpl := &template.Template{}
	tpl, err := tpl.Parse(g.StepCommand)
	if err != nil {
		return "", err
	}

	// TODO(raskyld): migrate this to an interface `KoEntrypointResolver` following the Strategy Pattern.
	//                Each resolver should be registered with a ko version range
	koAppName := path.Base(pkg.PkgPath)
	if koAppName == "." || koAppName == string(filepath.Separator) {
		koAppName = "ko-app"
	}

	args := stepCommandArgs{
		PkgName:    pkg.Name,
		ImportPath: pkg.PkgPath,
		KoAppName:  koAppName,
	}

	var command bytes.Buffer
	err = tpl.Execute(&command, args)
	if err != nil {
		return "", err
	}
	return command.String(), nil
}

func (g TaskYamlGenerator) buildWorkspaces(task unstructured.Unstructured, pkgMarkers markers.MarkerValues) error {
	if workspaces, ok := pkgMarkers[ttmarkers.MarkerWorkspace]; ok {
		workspacesYaml := make([]interface{}, 0, len(workspaces))
		for _, workspace := range workspaces {
			if workspace, isWorkspace := workspace.(ttmarkers.Workspace); isWorkspace {
				g.Logger.Info("found workspace", "workspace", workspace.Name)
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
	return nil
}

func (g TaskYamlGenerator) initTask(taskMarker interface{}) (unstructured.Unstructured, error) {
	var task unstructured.Unstructured
	if taskMarker, ok := taskMarker.(ttmarkers.Task); ok {
		task.SetName(taskMarker.Name)
		task.SetLabels(map[string]string{
			KubernetesVersionLabel: taskMarker.Version,
		})
	} else {
		return unstructured.Unstructured{}, errors.New("unexpected wrong type for task marker")
	}

	// For now we only support the v1 api
	task.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "tekton.dev",
		Version: "v1",
		Kind:    "Task",
	})
	return task, nil
}

// buildPackageDoc concatenates every package-level GoDoc to populate a Task description
func (g TaskYamlGenerator) buildPackageDoc(task unstructured.Unstructured, pkg *loader.Package) error {
	packagesDoc := make([]string, 0, len(pkg.Syntax))
	for _, file := range pkg.Syntax {
		if file != nil && file.Doc != nil {
			packagesDoc = append(packagesDoc, file.Doc.Text())
		}
	}

	if len(packagesDoc) != 0 {
		err := unstructured.SetNestedField(task.Object, strings.Join(packagesDoc, "\n"), "spec", "description")
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
