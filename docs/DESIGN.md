# Design

## Commands

### `init`

#### Description

Users can scaffold a new project.

#### Arguments

* `task_name`: The name of the Task, it will be used for the manifest `metadata.name`
* `dir`: Working directory in which to start the new project

#### Result

* Start a new Go project using a template (TBD) in the specified `dir`.
* In the template add the marker `tektasker:task:name=task_name` at the
  package-level.
