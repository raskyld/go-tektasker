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
	ttmarkers "github.com/raskyld/go-tektasker/pkg/markers"
	"log/slog"
	"sigs.k8s.io/controller-tools/pkg/genall"
	"sigs.k8s.io/controller-tools/pkg/markers"
	"strings"
	"text/template"
)

type TaskGoInternalGenerator struct {
	Logger      *slog.Logger
	Template    *template.Template
	PackageName string
	HeaderFile  string
	Year        string
}

func NewGoInternal(logger *slog.Logger, pkgName, headerFile, year string) (*TaskGoInternalGenerator, error) {
	g := &TaskGoInternalGenerator{
		Logger:      logger.With("generator", "goInternal"),
		Template:    &template.Template{},
		PackageName: pkgName,
		HeaderFile:  headerFile,
		Year:        year,
	}

	g.RegisterTemplate(GoHeaderName, GoHeaderTpl).
		RegisterTemplate(ParameterTypeName, ParameterTypeTpl).
		RegisterTemplate(ResultTypeName, ResultTypeTpl)

	return g, nil
}

func (g *TaskGoInternalGenerator) RegisterTemplate(name string, tpl string) *TaskGoInternalGenerator {
	nt, err := g.Template.New(name).Parse(tpl)
	if err != nil {
		panic(err)
	}

	g.Template = nt
	return g
}

func (*TaskGoInternalGenerator) RegisterMarkers(into *markers.Registry) error {
	return ttmarkers.Register(into)
}

func (g *TaskGoInternalGenerator) Generate(ctx *genall.GenerationContext) error {
	var headerBytes bytes.Buffer
	var headerText string

	if g.HeaderFile != "" {
		buf, err := ctx.ReadFile(g.HeaderFile)
		if err != nil {
			return err
		}

		headerText = strings.ReplaceAll(string(buf), " YEAR", " "+g.Year)
	}

	headerArgs := GoHeaderArgs{
		PkgName: g.PackageName,
		Header:  headerText,
		ImportPaths: []string{
			"errors",
			"fmt",
			"os",
			"strings",
		},
	}

	g.Logger.Debug("generating header", "args", headerArgs)
	err := g.Template.ExecuteTemplate(&headerBytes, GoHeaderName, headerArgs)
	if err != nil {
		return err
	}

	headerBytes.WriteByte('\n')

	g.Logger.Info("generating file", "file", "result.go")
	err = g.generatePkgFile(ctx, headerBytes, "result.go", ResultTypeName)
	if err != nil {
		return err
	}

	g.Logger.Info("generating file", "file", "parameter.go")
	err = g.generatePkgFile(ctx, headerBytes, "parameter.go", ParameterTypeName)
	if err != nil {
		return err
	}

	return nil
}

func (g *TaskGoInternalGenerator) generatePkgFile(ctx *genall.GenerationContext, headerBytes bytes.Buffer, fileName string, tplName string) error {
	output, err := ctx.OutputRule.Open(nil, fileName)
	if err != nil {
		return err
	}

	_, err = headerBytes.WriteTo(output)
	if err != nil {
		return err
	}

	err = g.Template.ExecuteTemplate(output, tplName, nil)
	if err != nil {
		return err
	}
	return nil
}
