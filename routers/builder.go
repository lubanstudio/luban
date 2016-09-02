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

func Builders(ctx *context.Context) {
	ctx.Data["Title"] = "Builders"

	builders, err := models.ListBuilders()
	if err != nil {
		ctx.Handle(500, "ListBuilders", err)
		return
	}
	ctx.Data["Builders"] = builders

	ctx.HTML(200, "builder/list")
}

func NewBuilder(ctx *context.Context) {
	ctx.Data["Title"] = "New Builder"
	ctx.HTML(200, "builder/new")
}

func NewBuilderPost(ctx *context.Context, form form.NewBuilder) {
	ctx.Data["Title"] = "New Builder"

	if ctx.HasError() {
		ctx.HTML(200, "builder/new")
		return
	}

	builder, err := models.NewBuilder(form.Name)
	if err != nil {
		if models.IsErrBuilderExists(err) {
			ctx.Data["Err_Name"] = true
			ctx.RenderWithErr("Builder name has been used.", "builder/new", form)
		} else {
			ctx.Handle(500, "NewBuilder", err)
		}
		return
	}

	ctx.Redirect(fmt.Sprintf("/builders/%d/edit", builder.ID))
}

func parseBuilderParams(ctx *context.Context) *models.Builder {
	builder, err := models.GetBuilderByID(ctx.ParamsInt64(":id"))
	if err != nil {
		if models.IsErrRecordNotFound(err) {
			ctx.NotFound()
		} else {
			ctx.Handle(500, "GetBuilderByID", err)
		}
		return nil
	}
	return builder
}

func EditBuilder(ctx *context.Context) {
	builder := parseBuilderParams(ctx)
	if ctx.Written() {
		return
	}
	ctx.Data["Builder"] = builder

	ctx.Data["Title"] = builder.Name + " - Builder"
	ctx.HTML(200, "builder/edit")
}

func EditBuilderPost(ctx *context.Context, form form.NewBuilder) {
	builder := parseBuilderParams(ctx)
	if ctx.Written() {
		return
	}
	ctx.Data["Builder"] = builder

	if ctx.HasError() {
		ctx.HTML(200, "builder/edit")
		return
	}

	builder.Name = form.Name
	builder.TrustLevel = models.ParseTrustLevel(form.TrustLevel)
	if err := builder.Save(); err != nil {
		if models.IsErrBuilderExists(err) {
			ctx.Data["Err_Name"] = true
			ctx.RenderWithErr("Builder name has been used.", "builder/edit", form)
		} else {
			ctx.Handle(500, "builder.Save", err)
		}
		return
	}

	ctx.Redirect(fmt.Sprintf("/builders/%d/edit", builder.ID))
}

func RegenerateBuilderToken(ctx *context.Context) {
	if err := models.RegenerateBuilderToken(ctx.ParamsInt64(":id")); err != nil {
		ctx.Handle(500, "RegenerateBuilderToken", err)
		return
	}

	ctx.Redirect(fmt.Sprintf("/builders/%d/edit", ctx.ParamsInt64(":id")))
}

func DeleteBuilder(ctx *context.Context) {
	if err := models.DeleteBuilderByID(ctx.ParamsInt64(":id")); err != nil {
		ctx.Handle(500, "DeleteBuilderByID", err)
		return
	}

	ctx.Redirect("/builders")
}
