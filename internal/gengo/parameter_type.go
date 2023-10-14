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

const ParameterTypeName = "parameter.type"

const ParameterTypeTpl = `// Parameter is capable of unmarshaling a specific Tekton parameter
type Parameter interface {
	// Unmarshal must result in the receiver being populated
	// from the passed bytes buffer
	Unmarshal([]byte) error

	// Name is the name of the parameter as it appears in your Task
	// manifest
	Name() string
}

// Read a parameter from environment variable or returns an error
func Read(v Parameter) error {
	envVarName := "PARAM_" + strings.ToUpper(v.Name()) + "_VALUE"

	envVarValue, ok := os.LookupEnv(envVarName)
	if !ok {
		return errors.New(fmt.Sprintf("parameter %s is not in environment (%s is missing)", v.Name(), envVarName))
	}

	err := v.Unmarshal([]byte(envVarValue))
	if err != nil {
		return err
	}

	return nil
}

// MustRead is like Read but will panic if it fails
func MustRead(v Parameter) {
	err := Read(v)
	if err != nil {
		panic(err.Error())
	}
}
`
