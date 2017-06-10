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
	"github.com/lubanstudio/luban/models"
	"github.com/lubanstudio/luban/pkg/context"
)

func Dashboard(ctx *context.Context) {
	ctx.Data["Title"] = "Dashboard"
	ctx.Data["PageIsDashboard"] = true

	ctx.Data["NumBuilders"] = models.CountBuilders()
	ctx.Data["NumTasks"] = models.CountTasks()

	ctx.HTML(200, "dashboard")
}
