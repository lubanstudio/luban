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

package routes

import (
	"fmt"

	"github.com/lubanstudio/luban/models"
	"github.com/lubanstudio/luban/pkg/context"
	"github.com/lubanstudio/luban/pkg/form"
	"github.com/lubanstudio/luban/pkg/setting"
)

func Tasks(c *context.Context) {
	c.Data["Title"] = "Tasks"

	tasks, err := models.ListTasks(1, 30)
	if err != nil {
		c.Handle(500, "ListTasks", err)
		return
	}
	c.Data["Tasks"] = tasks

	c.HTML(200, "task/list")
}

func NewTask(c *context.Context) {
	c.Data["Title"] = "New Task"
	form.AssignForm(form.NewTask{}, c.Data)
	c.HTML(200, "task/new")
}

func NewTaskPost(c *context.Context, form form.NewTask) {
	c.Data["Title"] = "New Task"

	if c.HasError() {
		c.HTML(200, "task/new")
		return
	}

	task, err := models.NewTask(c.User.ID, form.OS, form.Arch, form.Tags, form.Branch)
	if err != nil {
		if models.IsErrNoSuitableMatrix(err) {
			c.Data["Err_OS"] = true
			c.Data["Err_Arch"] = true
			c.Data["Err_Tags"] = true
			c.RenderWithErr(fmt.Sprintf("Fail to create task: %v", err), "task/new", form)
		} else {
			c.Handle(500, "NewTask", err)
		}
		return
	}

	c.Redirect(fmt.Sprintf("/tasks/%d", task.ID))
}

func NewBatchTasks(c *context.Context) {
	c.Data["Title"] = "New Batch Tasks"
	c.Data["branch"] = "master"
	c.HTML(200, "task/new_batch")
}

func NewBatchTasksPost(c *context.Context) {
	if err := models.NewBatchTasks(c.User.ID, c.Query("branch")); err != nil {
		c.Flash.Error("NewBatchTasks: " + err.Error())
	}
	c.Redirect("/tasks")
}

func ViewTask(c *context.Context) {
	c.Data["Title"] = c.Task.ID
	c.Data["PackFormats"] = setting.Project.PackFormats
	c.HTML(200, "task/view")
}

func ArchiveTask(c *context.Context) {
	if err := c.Task.Archive(); err != nil {
		c.RenderWithErr(fmt.Sprintf("Fail to archive task: %v", err), "task/view", nil)
		return
	}

	c.Redirect(fmt.Sprintf("/tasks/%d", c.Task.ID))
}
