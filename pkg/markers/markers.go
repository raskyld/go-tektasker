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
	MarkerParam  = "tektasker:param"
	MarkerResult = "tektasker:result"
	MarkerTask   = "tektasker:task"
)

type documentedMarker struct {
	*markers.Definition
	help *markers.DefinitionHelp
}

type hasHelp interface {
	Help() *markers.DefinitionHelp
}

// +controllertools:marker:generateHelp:category=task

// Param marks this struct as a parameter which can then be used as the
// target of an Unmarshalling operation
type Param struct {
	// Name is the name of your parameter
	Name string `marker:"name"`

	// Default is the default value you wish to set your parameter at if unspecified
	Default *string `marker:",optional"`

	Strict bool `marker:",optional"`
}

// +controllertools:marker:generateHelp:category=task

// Task marks this executable package as a runnable task for Tekton
type Task struct {
	// Name is the name of your parameter
	Name string `marker:"name"`

	Version string `marker:"version"`
}

// +controllertools:marker:generateHelp:category=task

// Result marks this struct as a result which can then be used as the
// target of a Marshalling operation
type Result struct {
	// Name is the name of the result
	Name string `marker:"name"`
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
