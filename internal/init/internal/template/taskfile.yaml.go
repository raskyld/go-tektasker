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
  default:
    silent: true
    ignore_error: true
    cmds:
      - cmd: "task -l"
      - cmd: "echo '\n\tStatus:\n'"
      - task: status

  status:
    desc: Check the version of the tools needed to build a Task and on which Kubernetes Cluster you are
    silent: true
    cmds:
      - "echo 'kubectl version:'"
      - "{{Raw ".PATH_KUBECTL"}} version --client | sed 's/^/\t/'"
      - "echo 'kubectl context where to apply:'"
      - |
          {{Raw "- if .APPLY_CONTEXT -"}}
          echo 'Fixed context: {{Raw ".APPLY_CONTEXT"}}' | sed 's/^/\t/'
          {{Raw "- else -"}}
          echo 'WARNING! It is recommended to set APPLY_CONTEXT in your .env file to avoid accidentally applying your task to the wrong cluster' && \
          {{Raw ".PATH_KUBECTL"}} config current-context | sed 's/^/\t/'
          {{Raw "- end -"}}
      - "echo 'ko builder version:'"
      - "{{Raw ".PATH_KO"}} version | sed 's/^/\t/'"
      - "echo 'tektasker version:'"
      - "{{Raw ".PATH_TEKTASKER"}} version | sed 's/^/\t/'"
      - "echo 'destination for manifests:'"
      - "echo '\tOUTPUT_MANIFESTS={{Raw ".OUTPUT_MANIFESTS"}}'"
      - "echo 'kustomize overlay deployed by task apply:'"
      - "echo '\tAPPLY_OVERLAY={{Raw ".APPLY_OVERLAY"}}'"

  manifest:
    desc: Generate your Task manifest as a Kustomization
    cmds:
      - "{{Raw ".PATH_TEKTASKER"}} generate manifest {{Raw ".OUTPUT_MANIFESTS"}}"

  generate:
    desc: Generate Go code for your project (run it everytime you change markers)
    cmds:
      - "{{Raw ".PATH_TEKTASKER"}} generate go {{Raw ".OUTPUT_INTERNAL_PKGS"}} {{Raw ".OUTPUT_INTERNAL_PKG_NAME"}}"

  apply:
    deps: ["generate", "manifest"]
    desc: Apply the changes on your current Kubernetes context
    cmds:
      - cmd: |
          {{Raw ".PATH_KUBECTL"}} kustomize {{Raw ".OUTPUT_MANIFESTS"}}/{{Raw ".APPLY_OVERLAY"}} | \
          {{Raw ".PATH_KO"}} resolve -f - | \
          {{Raw ".PATH_KUBECTL"}} apply -f -{{Raw "if .APPLY_CONTEXT"}} --context {{Raw ".APPLY_CONTEXT"}}{{Raw "end"}}
`
