// Copyright 2016 Unknwon
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package models

import (
	"fmt"
	"strings"
)

type ErrBuilderExists struct {
	Name string
}

func IsErrBuilderExists(err error) bool {
	_, ok := err.(ErrBuilderExists)
	return ok
}

func (err ErrBuilderExists) Error() string {
	return fmt.Sprintf("Builder already exists [name: %s]", err.Name)
}

type ErrNoSuitableMatrix struct {
	OS   string
	Arch string
	Tags []string
}

func IsErrNoSuitableMatrix(err error) bool {
	_, ok := err.(ErrNoSuitableMatrix)
	return ok
}

func (err ErrNoSuitableMatrix) Error() string {
	return fmt.Sprintf("no suitable matrix for the task [os: %s, arch: %s, tags: %s]", err.OS, err.Arch, strings.Join(err.Tags, ","))
}
