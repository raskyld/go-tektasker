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

const ResultTypeName = "result.type"

const ResultTypeTpl = `// Result allows you to simply write results back to Tekton
// for use in other tasks
type Result interface {
	// Marshal your type in any string representation.
	// JSON is recommended but Tekton does not enforce any format
	Marshal() ([]byte, error)

	// Name is the name of your Result as it appears in your Tekton
	// Task manifest
	Name() string
}

// Write the Result back to the filesystem for Tekton to consume it
func Write(r Result) error {
	envVarName := "RESULT_" + strings.ToUpper(r.Name()) + "_PATH"

	resultPath, ok := os.LookupEnv(envVarName)
	if !ok {
		return errors.New(fmt.Sprintf("result %s path could not be loaded (%s is missing)", r.Name(), envVarName))
	}

	resultValue, err := r.Marshal()
	if err != nil {
		return err
	}

	return os.WriteFile(resultPath, resultValue, 0666)
}

// MustWrite is like Write but will panic if it fails
func MustWrite(r Result) {
	err := Write(r)
	if err != nil {
		panic(err.Error())
	}
}
`
