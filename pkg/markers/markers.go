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

package markers

import (
	"sigs.k8s.io/controller-tools/pkg/markers"
)

var markersDef []documentedMarker

const (
	MarkerParam     = "tektasker:param"
	MarkerResult    = "tektasker:result"
	MarkerTask      = "tektasker:task"
	MarkerWorkspace = "tektasker:workspace"
)

type documentedMarker struct {
	*markers.Definition
	help *markers.DefinitionHelp
}

type hasHelp interface {
	Help() *markers.DefinitionHelp
}

// +controllertools:marker:generateHelp:category=task

// Param marks structs as Task parameter which can then be used in your code
// to take input from your users
type Param struct {
	// Name is the name of your parameter
	Name string `marker:"name"`

	// Default is the default value of the parameter
	Default *string `marker:",optional"`

	// Strict means you expect the parameter to strictly respect the format
	// of your struct. For this to be possible, the value passed to this parameter
	// by your user will need to be a valid JSON value that can be
	// unmarshalled into your struct, that's why you need to put valid JSON tags
	// in your structure fields.
	Strict bool `marker:",optional"`
}

// +controllertools:marker:generateHelp:category=task

// Task marks your package as a Task.
// Your package need to be executable to be bundled inside a container image,
// so you should use this marker on your main package
type Task struct {
	// Name is the name of your Task.
	// It will be used as the name of your Task manifest.
	Name string `marker:"name"`

	// Version is a way to communicate the version of your task to
	// your users
	Version string `marker:"version"`
}

// +controllertools:marker:generateHelp:category=task

// Result marks this struct as a result which means it can
// be Marshaled to populate the associated result
type Result struct {
	// Name is the name of the result
	Name string `marker:"name"`
}

// +controllertools:marker:generateHelp:category=task

// Workspace asks a workspace for this task
type Workspace struct {
	// Name is the name of the workspace
	Name string `marker:"name"`

	// Description for your workspace
	Description string `marker:"description"`

	// MountPath is useful to chose where to mount your workspace and
	// is always relative to root (`/`)
	MountPath string `marker:"mountPath,optional"`

	// ReadOnly defines if your workspace should be Read-Only,
	// remembers Tekton recommends to only have a single writeable
	// workspace
	ReadOnly bool `marker:"readOnly,optional"`

	// Optional defines if your user can choose not to provide this
	// workspace
	Optional bool `marker:"optional,optional"`
}

func define(name string, targetType markers.TargetType, help hasHelp) {
	markersDef = append(markersDef, documentedMarker{
		markers.Must(markers.MakeDefinition(name, targetType, help)),
		help.Help(),
	})
}

func init() {
	define(MarkerParam, markers.DescribesType, Param{})
	define(MarkerResult, markers.DescribesType, Result{})
	define(MarkerTask, markers.DescribesPackage, Task{})
	define(MarkerWorkspace, markers.DescribesPackage, Workspace{})
}

// Register all the markers in passed markers.Registry
func Register(into *markers.Registry) error {
	for _, def := range markersDef {
		err := into.Register(def.Definition)
		if err != nil {
			return err
		}

		if def.help != nil {
			into.AddHelp(def.Definition, def.help)
		}
	}

	return nil
}
