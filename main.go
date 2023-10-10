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

package main

import (
	"github.com/Raskyld/go-tektasker/internal/genyaml"
	"log/slog"
	"os"
	"sigs.k8s.io/controller-tools/pkg/genall"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
	var yamlGenerator genall.Generator = genyaml.TaskYamlGenerator{logger}

	generators := genall.Generators{&yamlGenerator}

	runtime, err := generators.ForRoots("./examples/...")
	if err != nil {
		panic(err)
	}

	runtime.OutputRules = genall.OutputRules{
		Default: genall.OutputToStdout,
	}

	runtime.Run()
}
