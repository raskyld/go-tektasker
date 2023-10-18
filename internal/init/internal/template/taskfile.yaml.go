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

package template

const TaskfileName = "taskfile"
const TaskfileTpl = `version: "3"

tasks:
  status:
    desc: Check the version of the tools needed to build a Task and on which Kubernetes Cluster you are
    cmds:
      - kubectl version --client
      - kubectl config current-context
      - ko version
`
