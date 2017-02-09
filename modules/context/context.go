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

package context

import (
	"fmt"
	"strings"

	"github.com/go-macaron/oauth2"
	"github.com/go-macaron/session"
	log "gopkg.in/clog.v1"
	"gopkg.in/macaron.v1"

	"github.com/lubanstudio/luban/models"
	"github.com/lubanstudio/luban/modules/form"
)

type Context struct {
	*macaron.Context
	Flash   *session.Flash
	Session session.Store

	User    *models.User
	Builder *models.Builder
	Task    *models.Task
}

// HasError returns true if error occurs in form validation.
func (ctx *Context) HasError() bool {
	hasErr, ok := ctx.Data["HasError"]
	if !ok {
		return false
	}
	ctx.Data["FlashTitle"] = "Form Validation"
	ctx.Flash.ErrorMsg = ctx.Data["ErrorMsg"].(string)
	ctx.Data["Flash"] = ctx.Flash
	return hasErr.(bool)
}

// RenderWithErr used for page has form validation but need to prompt error to users.
func (ctx *Context) RenderWithErr(msg string, tpl string, userForm interface{}) {
	if userForm != nil {
		form.AssignForm(userForm, ctx.Data)
	}
	ctx.Data["FlashTitle"] = "Form Validation"
	ctx.Flash.ErrorMsg = msg
	ctx.Data["Flash"] = ctx.Flash
	ctx.HTML(200, tpl)
}

func NotFound(ctx *Context) {
	ctx.Data["Title"] = "Page Not Found"
	ctx.HTML(404, "status/404")
}

func (ctx *Context) NotFound() {
	NotFound(ctx)
}

// Handle handles and logs error by given status.
func (ctx *Context) Handle(status int, title string, err error) {
	if err != nil {
		log.Error(4, "%s: %v", title, err)
		if macaron.Env != macaron.PROD {
			ctx.Data["ErrorMsg"] = err
		}
	}

	switch status {
	case 500:
		ctx.Data["Title"] = "Internal Server Error"
	}
	ctx.HTML(status, fmt.Sprintf("status/%d", status))
}

func Contexter() macaron.Handler {
	return func(c *macaron.Context, sess session.Store, f *session.Flash, tokens oauth2.Tokens) {
		ctx := &Context{
			Context: c,
			Flash:   f,
			Session: sess,
		}
		c.Map(ctx)

		ctx.Data["Link"] = strings.TrimSuffix(ctx.Req.URL.Path, "/")

		if ctx.Session.Get(oauth2.KEY_TOKEN) != nil {
			user, err := models.GetOrCreateUserByOAuthID(tokens.Access())
			if err != nil {
				ctx.Handle(500, "GetOrCreateUserByOAuthID", err)
				return
			}
			ctx.User = user
			ctx.Data["IsSigned"] = true
			ctx.Data["User"] = user

			log.Trace("Authenticated user: %s", user.Username)
		}
	}
}
