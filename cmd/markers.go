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
	"errors"
	ttmarkers "github.com/raskyld/go-tektasker/pkg/markers"
	"github.com/spf13/cobra"
	"os"
	"sigs.k8s.io/controller-tools/pkg/genall/help"
	"sigs.k8s.io/controller-tools/pkg/genall/help/pretty"
	"sigs.k8s.io/controller-tools/pkg/markers"
	"strings"
)

func NewMarkers(ctx *Context) *cobra.Command {
	markersCmd := &cobra.Command{
		Use:   "markers [marker] [package | type]",
		Args:  cobra.MaximumNArgs(2),
		Short: "Show help for all markers you can use in your code or for the one specified",
		RunE: func(cmd *cobra.Command, args []string) error {
			registry := &markers.Registry{}

			err := ttmarkers.Register(registry)
			if err != nil {
				return err
			}

			if len(args) > 0 {
				markerName := args[0]
				typeName := "type"
				if len(args) == 2 {
					typeName = args[1]
				}

				// Just to help the user when they forget the "+"
				if len(markerName) > 0 && markerName[0] != '+' {
					markerName = "+" + markerName
				}

				targetType := markers.DescribesType

				switch strings.ToLower(typeName) {
				case "package":
					targetType = markers.DescribesPackage
				case "field":
					targetType = markers.DescribesField
				}

				return showMarker(registry, markerName, targetType)
			} else {
				return showAllMarkers(registry)
			}
		},
	}

	return markersCmd
}

func showMarker(registry *markers.Registry, name string, targetType markers.TargetType) error {
	def := registry.Lookup(name, targetType)
	if def == nil {
		return errors.New("marker not found")
	}

	doc := help.ForDefinition(def, registry.HelpFor(def))
	err := pretty.MarkersDetails(true, doc.Category, []help.MarkerDoc{doc}).WriteTo(os.Stdout)

	return err
}

func showAllMarkers(registry *markers.Registry) error {
	docByCategory := help.ByCategory(registry, help.SortByCategory)
	for _, doc := range docByCategory {
		err := pretty.MarkersSummary(doc.Category, doc.Markers).WriteTo(os.Stdout)
		if err != nil {
			return err
		}
	}

	return nil
}
