# Go Tektasker

> :warning: This is a WIP and the main branch will stay non-functional
> for a while, we are really early in the project ;)

A framework for building
[Tekton](https://tekton.dev)
[Tasks](https://tekton.dev/docs/pipelines/tasks/) in Go.

## :wrench: State of the `main` branch

As of now, the CLI is capable of generating manifests for a list of Go package
and outputting them in a directory.
* The GVK is not appearing, I'm investigating...
* The steps are not produced yet, I am thinking about producing a `base`
  Kustomization for every Task with a single step that is the Go program
  the user is making
* In this single-step, every parameter should be passed as an environment var

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
