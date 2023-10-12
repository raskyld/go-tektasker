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

// Package examples is a Tekton Task:
//
// Print a Hello World sorting people by score.
package examples

// +tektasker:task:name=example,version=0.1

// +tektasker:param:name=msg

// Message is the message you want to send to your user
type Message string

// +tektasker:param:name=names,default="[\"jeremy\", \"virginie\"]"

// Names are the names that will be used in the Hello World!
type Names []string

// +tektasker:param:name=scores,default="{\"jeremy\": \"10\", \"virginie\": \"10\"}"

// Scores maps a name to a score
type Scores map[string]string

// +tektasker:param:name=sscore,default="{\"name\": \"virginie\", \"score\": \"10\"}"

// SingleScore is the score of a single person
type SingleScore struct {
	Name  string
	Score int
}

// +tektasker:param:name=strictscore,default="{\"name\": \"jeremy\", \"score\": \"10\"}",strict=true

// StrictScore is the score of a single person
type StrictScore struct {
	Name  string `json:"name"`
	Score int    `json:"score"`
}

// +tektasker:result:name=structresult

// StructResult is a result that will be marshaled to valid JSON
type StructResult struct{}

// +tektasker:result:name=arrayresult

// ArrayResult is a result that will be marshaled to a valid JSON array value
type ArrayResult []interface{}
