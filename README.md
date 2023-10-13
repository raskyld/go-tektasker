# Go Tektasker

> :warning: This is a WIP and the main branch will stay non-functional
> for a while, we are really early in the project ;)

A framework for building
[Tekton](https://tekton.dev)
[Tasks](https://tekton.dev/docs/pipelines/tasks/) in Go.

## :wrench: State of the `main` branch

As of now, the CLI is capable of generating manifests for a list of Go package
and outputting them in a directory or stdout.

This is the resulting YAML manifest of running tektasker on the package in
`examples/`:

```yaml
---
apiVersion: tekton.dev/v1
kind: Task
metadata:
  labels:
    app.kubernetes.io/version: "0.1"
  name: example
spec:
  description: |
    Package examples is a Tekton Task:

    Print a Hello World sorting people by score.
  params:
    - description: Message is the message you want to send to your user
      name: msg
      type: string
    - default:
        - jeremy
        - virginie
      description: Names are the names that will be used in the Hello World!
      name: names
      type: array
    - default: '{"jeremy": "10", "virginie": "10"}'
      description: Scores maps a name to a score
      name: scores
      type: string
    - default: '{"name": "virginie", "score": "10"}'
      description: SingleScore is the score of a single person
      name: sscore
      type: string
    - default:
        name: jeremy
        score: "10"
      description: StrictScore is the score of a single person
      name: strictscore
      properties:
        name:
          type: string
        score:
          type: string
      type: object
  results:
    - description: StructResult is a result that will be marshaled to valid JSON
      name: structresult
      type: string
    - description: ArrayResult is a result that will be marshaled to a valid JSON array
        value
      name: arrayresult
      type: array
  workspaces:
    - description: first workspace
      name: workspace1
      optional: false
      readOnly: false
    - description: second workspace
      name: workspace2
      optional: false
      readOnly: true
    - description: third workspace
      name: workspace3
      optional: true
      readOnly: true
```

## Road-map

* Scaffold a simple Go project that will serves as the Task codebase
* Adding `+tektasker:param` on a struct marks it as a
  [parameter](https://tekton.dev/docs/pipelines/tasks/#specifying-parameters)
* Adding `+tektasker:result` on a struct marks it as a
  [result](https://tekton.dev/docs/pipelines/tasks/#emitting-results)
* Provide some helper functions to Marshal/Unmarshal Results and Parameters
* Generate the YAML manifest of the resulting
  [task](https://tekton.dev/docs/pipelines/tasks/#configuring-a-task)
* Build and push the image on the behalf of
  the user using [ko](https://github.com/ko-build/ko)

## Contributing

See [CONTRIBUTING.md](./CONTRIBUTING.md)
