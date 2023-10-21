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

const TaskGoName = "task.go"
const TaskGoTpl = `package main

import (
    "fmt"
)

// +tektasker:task:name={{.TaskName}},version=0.1

// Thanks a lot for using Tektasker! <3

// +tektasker:param:name=lover1,default=tekton

// LoverOne is the person LoverTwo is in love with
type LoverOne string

// +tektasker:param:name=lover2,default=tektasker

// LoverTwo is the person LoverOne is in love with
type LoverTwo string

// +tektasker:result:name=lovemaking

// LoveMaking will hold the result of the concatenation
type LoveMaking string

func main() {
    // Allocate memory for the params of the task
    var l1 LoverOne
    var l2 LoverTwo

    // Allocate memory for the results of the task
    var lm LoveMaking

    // You then need to run "task generate" to get a copy of
    // Tektasker helper code, once it is done, you can uncomment
    // the following lines:
    
    // tekton.MustRead(&l1)
    // tekton.MustRead(&l2)

    // Here we create a message from the two parameter of the task
    lm = LoveMaking(fmt.Sprintf("%s + %s = love!", l1, l2))

    // The following line will persist the result back to Tekton!

    // tekton.MustWrite(&lm)

    // Well done! Now, you can run "task apply" to write your task to your
    // current Kubernetes context and create a TaskRun that references the just
    // applied Task named "{{.TaskName}}" to launch it on your Tekton cluster

    // If you use the default values of the parameter, you should see the result
    // "tekton + tektasker = love!" in your TaskRun once it is executed!

    // Have fun with Tektasker and please report any issue back to the project :)
}
`

type TaskGoArgs struct {
	TaskName string
}
