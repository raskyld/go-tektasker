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

func TestParamFuncName(t *testing.T) {
	tpl, err := template.New(ParamFuncNameName).Parse(ParamFuncNameTpl)
	if err != nil {
		t.Errorf("couldnt create template %s: %s", ParamFuncNameName, err.Error())
	}

	tests := []struct {
		name    string
		args    ParamFuncArgs
		wantErr bool
		result  string
	}{
		{
			"Simple param",
			ParamFuncArgs{
				ParamName: "param1",
				ParamType: "ParamOne",
			},
			false,
			`func (param *ParamOne) Name() string {
	return "param1"
}
`,
		},
	}

	var buffer bytes.Buffer
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			defer buffer.Reset()
			err := tpl.ExecuteTemplate(&buffer, ParamFuncNameName, test.args)
			if test.wantErr && err == nil {
				t.Error("should have failed")
			}

			if !reflect.DeepEqual(buffer.String(), test.result) {
				t.Errorf("unwanted diff, got\n---\n%s\n---\nwanted\n---\n%s", buffer.String(), test.result)
			}
		})
	}
}

func TestParamFuncUnmarshalSimple(t *testing.T) {
	tpl, err := template.New(ParamFuncUnmarshalSimpleName).Parse(ParamFuncUnmarshalSimpleTpl)
	if err != nil {
		t.Errorf("couldnt create template %s: %s", ParamFuncUnmarshalSimpleName, err.Error())
	}

	tests := []struct {
		name    string
		args    ParamFuncArgs
		wantErr bool
		result  string
	}{
		{
			"Simple param",
			ParamFuncArgs{
				ParamName: "param1",
				ParamType: "ParamOne",
			},
			false,
			`func (param *ParamOne) Unmarshal(buf []byte) error {
	*param = ParamOne(buf)
	return nil
}
`,
		},
	}

	var buffer bytes.Buffer
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			defer buffer.Reset()
			err := tpl.ExecuteTemplate(&buffer, ParamFuncUnmarshalSimpleName, test.args)
			if test.wantErr && err == nil {
				t.Error("should have failed")
			}

			if !reflect.DeepEqual(buffer.String(), test.result) {
				t.Errorf("unwanted diff, got\n---\n%s\n---\nwanted\n---\n%s", buffer.String(), test.result)
			}
		})
	}
}

func TestParamFuncUnmarshalJSON(t *testing.T) {
	tpl, err := template.New(ParamFuncUnmarshalJSONName).Parse(ParamFuncUnmarshalJSONTpl)
	if err != nil {
		t.Errorf("couldnt create template %s: %s", ParamFuncUnmarshalJSONName, err.Error())
	}

	tests := []struct {
		name    string
		args    ParamFuncArgs
		wantErr bool
		result  string
	}{
		{
			"Simple param",
			ParamFuncArgs{
				ParamName: "param1",
				ParamType: "ParamOne",
			},
			false,
			`func (param *ParamOne) Unmarshal(buf []byte) error {
	return json.Unmarshal(buf, param)
}
`,
		},
	}

	var buffer bytes.Buffer
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			defer buffer.Reset()
			err := tpl.ExecuteTemplate(&buffer, ParamFuncUnmarshalJSONName, test.args)
			if test.wantErr && err == nil {
				t.Error("should have failed")
			}

			if !reflect.DeepEqual(buffer.String(), test.result) {
				t.Errorf("unwanted diff, got\n---\n%s\n---\nwanted\n---\n%s", buffer.String(), test.result)
			}
		})
	}
}
