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

const ParamFuncNameName = "param.func.name"

const ParamFuncNameTpl = `func (param *{{.ParamType}}) Name() string {
	return "{{.ParamName}}"
}
`

const ParamFuncUnmarshalSimpleName = "param.func.unmarshal.simple"

const ParamFuncUnmarshalSimpleTpl = `func (param *{{.ParamType}}) Unmarshal(buf []byte) error {
	*param = {{.ParamType}}(buf)
	return nil
}
`

const ParamFuncUnmarshalJSONName = "param.func.unmarshal.json"

const ParamFuncUnmarshalJSONTpl = `func (param *{{.ParamType}}) Unmarshal(buf []byte) error {
	return json.Unmarshal(buf, param)
}
`

type ParamFuncArgs struct {
	ParamName string
	ParamType string
}
