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
	"fmt"
	ttmarkers "github.com/Raskyld/go-tektasker/pkg/markers"
	"sigs.k8s.io/controller-tools/pkg/genall"
	"sigs.k8s.io/controller-tools/pkg/markers"
)

type TaskYamlGenerator struct{}

func (TaskYamlGenerator) RegisterMarkers(into *markers.Registry) error {
	return ttmarkers.Register(into)
}

func (TaskYamlGenerator) Generate(ctx *genall.GenerationContext) error {
	for _, pkg := range ctx.Roots {
		// TODO(raskyld): get all package level markers

		err := markers.EachType(ctx.Collector, pkg, func(info *markers.TypeInfo) {
			for k, v := range info.Markers {
				fmt.Printf("Found a marker %s on %s, with value %+v\n", k, info.RawSpec.Name, v)
			}
		})

		if err != nil {
			return err
		}
	}

	return nil
}
