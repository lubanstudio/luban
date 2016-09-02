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
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/lubanstudio/luban/models"
	"github.com/lubanstudio/luban/modules/context"
	"github.com/lubanstudio/luban/modules/setting"
)

func RequireBuilderToken(ctx *context.Context) {
	var err error
	ctx.Builder, err = models.GetBuilderByToken(ctx.Req.Header.Get("X-LUBAN-TOKEN"))
	if err != nil {
		if models.IsErrRecordNotFound(err) {
			ctx.Status(403)
		} else {
			ctx.Error(500, fmt.Sprintf("GetBuilderByToken: %v", err))
		}
		return
	}
}

func UpdateMatrix(ctx *context.Context) {
	data, err := ctx.Req.Body().Bytes()
	if err != nil {
		ctx.Error(500, fmt.Sprintf("Req.Body().Bytes: %v", err))
		return
	}

	rawMatrices := make([]*setting.Matrix, 0, 3)
	if err = json.Unmarshal(data, &rawMatrices); err != nil {
		ctx.Error(500, fmt.Sprintf("json.Unmarshal: %v", err))
		return
	}

	matrices := make([]*models.Matrix, 0, 5)
	for _, raw := range rawMatrices {
		sort.Strings(raw.Tags)
		for _, arch := range raw.Archs {
			matrices = append(matrices, &models.Matrix{
				OS:   raw.OS,
				Arch: arch,
				Tags: strings.Join(raw.Tags, ","),
			})
		}
	}

	if err = ctx.Builder.UpdateMatrices(matrices); err != nil {
		ctx.Error(500, fmt.Sprintf("UpdateMatrices: %v", err))
		return
	}

	ctx.Status(204)
}

func HeartBeat(ctx *context.Context) {
	isIdle := ctx.Req.Header.Get("X-LUBAN-STATUS") == "IDLE"
	if isIdle {
		task, err := models.GetTaskByID(ctx.Builder.TaskID)
		if err != nil {
			ctx.Error(500, fmt.Sprintf("GetTaskByID: %v", err))
			return
		}
		ctx.JSON(200, map[string]interface{}{
			"clone_url": setting.Project.CloneURL,
			"task": map[string]interface{}{
				"id":     task.ID,
				"os":     task.OS,
				"arch":   task.Arch,
				"tags":   task.Tags,
				"commit": task.Commit,
			},
		})
	}

	if err := ctx.Builder.HeartBeat(isIdle); err != nil {
		ctx.Error(500, fmt.Sprintf("HeartBeat: %v", err))
		return
	}

	if ctx.Written() {
		return
	}
	ctx.Status(204)
}
