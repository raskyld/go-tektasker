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
	"reflect"
	"testing"
	"text/template"
)

func TestResultFuncName(t *testing.T) {
	tpl, err := template.New(ResultFuncNameName).Parse(ResultFuncNameTpl)
	if err != nil {
		t.Errorf("couldnt create template %s: %s", ResultFuncNameName, err.Error())
	}

	tests := []struct {
		name    string
		args    ResultFuncArgs
		wantErr bool
		result  string
	}{
		{
			"Simple result",
			ResultFuncArgs{
				ResultName: "result1",
				ResultType: "ResultOne",
			},
			false,
			`func (result *ResultOne) Name() string {
	return "result1"
}
`,
		},
	}

	var buffer bytes.Buffer
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			defer buffer.Reset()
			err := tpl.ExecuteTemplate(&buffer, ResultFuncNameName, test.args)
			if test.wantErr && err == nil {
				t.Error("should have failed")
			}

			if !reflect.DeepEqual(buffer.String(), test.result) {
				t.Errorf("unwanted diff, got\n---\n%s\n---\nwanted\n---\n%s", buffer.String(), test.result)
			}
		})
	}
}

func TestResultFuncMarshalSimple(t *testing.T) {
	tpl, err := template.New(ResultFuncMarshalSimpleName).Parse(ResultFuncMarshalSimpleTpl)
	if err != nil {
		t.Errorf("couldnt create template %s: %s", ResultFuncMarshalSimpleName, err.Error())
	}

	tests := []struct {
		name    string
		args    ResultFuncArgs
		wantErr bool
		result  string
	}{
		{
			"Simple result",
			ResultFuncArgs{
				ResultName: "result1",
				ResultType: "ResultOne",
			},
			false,
			`func (result *ResultOne) Marshal() ([]byte, error) {
	return []byte(*result), nil
}
`,
		},
	}

	var buffer bytes.Buffer
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			defer buffer.Reset()
			err := tpl.ExecuteTemplate(&buffer, ResultFuncMarshalSimpleName, test.args)
			if test.wantErr && err == nil {
				t.Error("should have failed")
			}

			if !reflect.DeepEqual(buffer.String(), test.result) {
				t.Errorf("unwanted diff, got\n---\n%s\n---\nwanted\n---\n%s", buffer.String(), test.result)
			}
		})
	}
}

func TestResultFuncMarshalJSON(t *testing.T) {
	tpl, err := template.New(ResultFuncMarshalJSONName).Parse(ResultFuncMarshalJSONTpl)
	if err != nil {
		t.Errorf("couldnt create template %s: %s", ResultFuncMarshalJSONName, err.Error())
	}

	tests := []struct {
		name    string
		args    ResultFuncArgs
		wantErr bool
		result  string
	}{
		{
			"Simple result",
			ResultFuncArgs{
				ResultName: "result1",
				ResultType: "ResultOne",
			},
			false,
			`func (result *ResultOne) Marshal() ([]byte, error) {
	return json.Marshal(result)
}
`,
		},
	}

	var buffer bytes.Buffer
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			defer buffer.Reset()
			err := tpl.ExecuteTemplate(&buffer, ResultFuncMarshalJSONName, test.args)
			if test.wantErr && err == nil {
				t.Error("should have failed")
			}

			if !reflect.DeepEqual(buffer.String(), test.result) {
				t.Errorf("unwanted diff, got\n---\n%s\n---\nwanted\n---\n%s", buffer.String(), test.result)
			}
		})
	}
}
