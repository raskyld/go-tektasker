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

import "sigs.k8s.io/controller-tools/pkg/markers"

var markersDef []documentedMarker

type documentedMarker struct {
	*markers.Definition
	help *markers.DefinitionHelp
}

type hasHelp interface {
	Help() *markers.DefinitionHelp
}

// +controllertools:marker:generateHelp:category=task

// Param tells Tektasker that this field can be used to receive the corresponding Task's param
type Param struct {
	// Name of your param
	Name string `marker:"name"`

	// Default value for your param
	Default *string `marker:",optional"`
}

func define(name string, targetType markers.TargetType, help hasHelp) {
	markersDef = append(markersDef, documentedMarker{
		markers.Must(markers.MakeDefinition(name, targetType, help)),
		help.Help(),
	})
}

func init() {
	// TODO(raskyld): generate Help() with helpgen
	//define("tektasker:param", markers.DescribesField, Param{})
}
