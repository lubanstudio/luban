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

package main

import (
	"fmt"
	"net/http"
	"path"

	"github.com/go-macaron/binding"
	"github.com/go-macaron/oauth2"
	"github.com/go-macaron/session"
	goauth2 "golang.org/x/oauth2"
	log "gopkg.in/clog.v1"
	"gopkg.in/macaron.v1"

	"github.com/lubanstudio/luban/models"
	"github.com/lubanstudio/luban/pkg/context"
	"github.com/lubanstudio/luban/pkg/form"
	"github.com/lubanstudio/luban/pkg/setting"
	"github.com/lubanstudio/luban/pkg/template"
	"github.com/lubanstudio/luban/routes"
)

const APP_VER = "0.6.0.0610"

func init() {
	setting.AppVer = APP_VER
}

func main() {
	log.Info("Luban %s", APP_VER)

	m := macaron.New()
	if !setting.ProdMode {
		m.Use(macaron.Logger())
	}
	m.Use(macaron.Recovery())
	m.Use(macaron.Static("public", macaron.StaticOptions{
		SkipLogging: setting.ProdMode,
	}))
	m.Use(macaron.Renderer(macaron.RenderOptions{
		Funcs:      template.NewFuncMap(),
		IndentJSON: macaron.Env != macaron.PROD,
	}))
	m.Use(session.Sessioner(session.Options{
		Provider:       "file",
		ProviderConfig: "data/sessions",
	}))
	m.Use(oauth2.Github(
		&goauth2.Config{
			ClientID:     setting.OAuth2.ClientID,
			ClientSecret: setting.OAuth2.ClientSecret,
		},
	))
	m.Use(context.Contexter())

	bindIgnErr := binding.BindIgnErr

	m.Get("/", func(ctx *macaron.Context) { ctx.Redirect("/dashboard") })
	m.Group("", func() {
		m.Get("/dashboard", routes.Dashboard)

		m.Group("/tasks", func() {
			m.Get("", routes.Tasks)

			m.Group("", func() {
				m.Combo("/new").Get(routes.NewTask).Post(bindIgnErr(form.NewTask{}), routes.NewTaskPost)
				m.Combo("/new_batch", context.ReqAdmin()).Get(routes.NewBatchTasks).Post(routes.NewBatchTasksPost)

				m.Group("/:id", func() {
					m.Get("", routes.ViewTask)
					m.Get("/archive", context.ReqAdmin(), routes.ArchiveTask)
				}, func(ctx *context.Context) {
					task, err := models.GetTaskByID(ctx.ParamsInt64(":id"))
					if err != nil {
						if models.IsErrRecordNotFound(err) {
							ctx.NotFound()
						} else {
							ctx.Handle(500, "GetTaskByID", err)
						}
						return
					}
					ctx.Task = task
					ctx.Data["Task"] = ctx.Task
				})
			})
		}, func(ctx *context.Context) {
			ctx.Data["PageIsTask"] = true
			ctx.Data["AllowedOSs"] = setting.AllowedOSs
			ctx.Data["AllowedArchs"] = setting.AllowedArchs
			ctx.Data["AllowedTags"] = setting.AllowedTags
			ctx.Data["AllowedBranches"] = setting.Project.Branches
		})

		m.Group("/builders", func() {
			m.Get("", routes.Builders)

			m.Group("", func() {
				m.Combo("/new").Get(routes.NewBuilder).Post(bindIgnErr(form.NewBuilder{}), routes.NewBuilderPost)

				m.Group("/:id", func() {
					m.Combo("/edit").Get(routes.EditBuilder).Post(bindIgnErr(form.NewBuilder{}), routes.EditBuilderPost)
					m.Post("/regenerate_token", routes.RegenerateBuilderToken)
					m.Post("/delete", routes.DeleteBuilder)
				})
			}, context.ReqAdmin())
		}, func(ctx *context.Context) {
			ctx.Data["PageIsBuilder"] = true
		})

	}, oauth2.LoginRequired)

	m.Get("/artifacts/:name", func(ctx *context.Context) {
		http.ServeFile(ctx.Resp, ctx.Req.Request, path.Join(setting.ArtifactsPath, ctx.Params(":name")))
	})

	m.Group("/api/v1", func() {
		m.Group("/builder", func() {
			m.Post("/matrix", routes.UpdateMatrix)
			m.Post("/heartbeat", routes.HeartBeat)
			m.Post("/upload/artifact", routes.UploadArtifact)
		}, routes.RequireBuilderToken)
	})

	m.NotFound(context.NotFound)

	go models.AssignTasks()

	listenAddr := fmt.Sprintf("0.0.0.0:%d", setting.HTTPPort)
	log.Info("Listening on %s", listenAddr)
	log.Fatal(4, "Fail to start server: %v", http.ListenAndServe(listenAddr, m))
}
