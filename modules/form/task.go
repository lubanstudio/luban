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

package form

import (
	"github.com/go-macaron/binding"
	"gopkg.in/macaron.v1"
)

type NewTask struct {
	OS     string `form:"os" binding:"Required"`
	Arch   string `binding:"Required"`
	Tags   []string
	Branch string `binding:"Required"`
}

func (f *NewTask) Validate(ctx *macaron.Context, errs binding.Errors) binding.Errors {
	return Validate(errs, ctx.Data, f)
}
