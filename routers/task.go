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

package routers

import (
	"fmt"

	"github.com/lubanstudio/luban/models"
	"github.com/lubanstudio/luban/modules/context"
	"github.com/lubanstudio/luban/modules/form"
)

func Tasks(ctx *context.Context) {
	ctx.Data["Title"] = "Tasks"

	tasks, err := models.ListTasks()
	if err != nil {
		ctx.Handle(500, "ListTasks", err)
		return
	}
	ctx.Data["Tasks"] = tasks

	ctx.HTML(200, "task/list")
}

func NewTask(ctx *context.Context) {
	ctx.Data["Title"] = "New Task"
	form.AssignForm(form.NewTask{}, ctx.Data)
	ctx.HTML(200, "task/new")
}

func NewTaskPost(ctx *context.Context, form form.NewTask) {
	ctx.Data["Title"] = "New Task"

	if ctx.HasError() {
		ctx.HTML(200, "task/new")
		return
	}

	task, err := models.NewTask(ctx.User.ID, form.OS, form.Arch, form.Tags, form.Branch)
	if err != nil {
		if models.IsErrNoSuitableMatrix(err) {
			ctx.Data["Err_OS"] = true
			ctx.Data["Err_Arch"] = true
			ctx.Data["Err_Tags"] = true
			ctx.RenderWithErr(fmt.Sprintf("Fail to create task: %v", err), "task/new", form)
		} else {
			ctx.Handle(500, "NewTask", err)
		}
		return
	}

	ctx.Redirect(fmt.Sprintf("/tasks/%d", task.ID))
}

func ViewTask(ctx *context.Context) {
	task, err := models.GetTaskByID(ctx.ParamsInt64(":id"))
	if err != nil {
		if models.IsErrRecordNotFound(err) {
			ctx.NotFound()
		} else {
			ctx.Handle(500, "GetTaskByID", err)
		}
		return
	}
	ctx.Data["Title"] = task.ID

	ctx.Data["Task"] = task
	ctx.HTML(200, "task/view")
}
