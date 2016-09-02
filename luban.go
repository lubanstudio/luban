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

	log "github.com/Sirupsen/logrus"
	"github.com/go-macaron/binding"
	"github.com/go-macaron/oauth2"
	"github.com/go-macaron/session"
	goauth2 "golang.org/x/oauth2"
	"gopkg.in/macaron.v1"

	"github.com/lubanstudio/luban/modules/context"
	"github.com/lubanstudio/luban/modules/form"
	"github.com/lubanstudio/luban/modules/setting"
	"github.com/lubanstudio/luban/modules/template"
	"github.com/lubanstudio/luban/routers"
)

const APP_VER = "0.3.2.0902"

func init() {
	setting.AppVer = APP_VER
}

func main() {
	log.Printf("Luban %s", APP_VER)

	m := macaron.Classic()
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
		m.Get("/dashboard", routers.Dashboard)

		m.Group("/tasks", func() {
			m.Get("", routers.Tasks)
		})

		m.Group("/builders", func() {
			m.Get("", routers.Builders)

			m.Group("", func() {
				m.Combo("/new").Get(routers.NewBuilder).Post(bindIgnErr(form.NewBuilder{}), routers.NewBuilderPost)

				m.Group("/:id", func() {
					m.Combo("/edit").Get(routers.EditBuilder).Post(bindIgnErr(form.NewBuilder{}), routers.EditBuilderPost)
					m.Post("/regenerate_token", routers.RegenerateBuilderToken)
					m.Post("/delete", routers.DeleteBuilder)
				})
			}, context.ReqAdmin())
		})
	}, oauth2.LoginRequired)

	m.Group("/api/v1", func() {
		m.Group("/builder", func() {
			m.Post("/matrix", routers.UpdateMatrix)
			m.Post("/heartbeat", routers.HeartBeat)
		}, routers.RequireBuilderToken)
	})

	m.NotFound(context.NotFound)

	listenAddr := fmt.Sprintf("0.0.0.0:%d", setting.HTTPPort)
	log.Println("Listening on", listenAddr)
	log.Fatal(http.ListenAndServe(listenAddr, m))
}
