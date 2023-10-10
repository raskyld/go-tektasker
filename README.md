# Go Tektasker

> :warning: This is a WIP and the main branch will stay non-functional
> for a while, we are really early in the project ;)

A framework for building
[Tekton](https://tekton.dev)
[Tasks](https://tekton.dev/docs/pipelines/tasks/) in Go.

## :wrench: State of the `main` branch

As of now, building and running this project will output the
result of parsing the package in `examples/`, this is very promising
but there is still a lot to do!

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
