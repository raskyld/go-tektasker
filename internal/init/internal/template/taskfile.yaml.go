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

dotenv:
  - .env

tasks:
  status:
    desc: Check the version of the tools needed to build a Task and on which Kubernetes Cluster you are
    cmds:
      - "{{Raw ".KUBECTL_PATH"}} version --client"
      - "{{Raw ".KUBECTL_PATH"}} config current-context"
      - "{{Raw ".KO_PATH"}} version"
      - "{{Raw ".TT_PATH"}} version"

  manifest:
    desc: Generate your Task manifest as a Kustomization
    cmds:
      - "{{Raw ".TT_PATH"}} generate manifest {{Raw ".TT_OUTPUT_MANIFEST"}}"

  generate:
    desc: Generate Go code for your project (run it everytime you change markers)
    cmds:
      - "{{Raw ".TT_PATH"}} generate go {{Raw ".PROJECT_INTERNAL_PKGS"}} {{Raw ".TT_INTERNAL_PKG_NAME"}}"

  apply:
    deps: ["generate", "manifest"]
    desc: Apply the changes on your current Kubernetes context
    cmds:
      - cmd: |
          {{Raw ".KUBECTL_PATH"}} kustomize {{Raw ".TT_OUTPUT_MANIFEST"}}/{{Raw ".PROJECT_TASK_OVERLAY"}} | \
          {{Raw ".KO_PATH"}} apply -f -
`
