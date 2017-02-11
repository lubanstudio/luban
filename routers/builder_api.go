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
	"io"
	"os"
	"path"
	"sort"
	"strings"

	log "gopkg.in/clog.v1"

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
			ctx.Error("GetBuilderByToken: %v", err)
		}
		return
	}
}

func UpdateMatrix(ctx *context.Context) {
	data, err := ctx.Req.Body().Bytes()
	if err != nil {
		ctx.Error("Req.Body().Bytes: %v", err)
		return
	}

	rawMatrices := make([]*setting.Matrix, 0, 3)
	if err = json.Unmarshal(data, &rawMatrices); err != nil {
		ctx.Error("json.Unmarshal: %v", err)
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
		ctx.Error("UpdateMatrices: %v", err)
		return
	}

	ctx.Status(204)
}

func HeartBeat(ctx *context.Context) {
	status := ctx.Req.Header.Get("X-LUBAN-STATUS")
	log.Trace("Hearrbeat from builder '%d': %s", ctx.Builder.ID, status)

	isIdle := status == "IDLE"
	if err := ctx.Builder.HeartBeat(isIdle && ctx.Builder.TaskID == 0); err != nil {
		log.Error(4, "HeartBeat [%d]: %v", ctx.Builder.ID, err)
		ctx.Error("HeartBeat: %v", err)
		return
	}

	if isIdle {
		// Response assgined task to builder if it's idle.
		if ctx.Builder.TaskID > 0 {
			isIdle = false
			task, err := models.GetTaskByID(ctx.Builder.TaskID)
			if err != nil {
				ctx.Error("GetTaskByID [%d]: %v", ctx.Builder.TaskID, err)
				return
			}

			ctx.Resp.Header().Set("X-LUBAN-TASK", "ASSIGN")
			ctx.JSON(200, map[string]interface{}{
				"import_path":  setting.Project.ImportPath,
				"pack_root":    setting.Project.PackRoot,
				"pack_entries": setting.Project.PackEntries,
				"pack_formats": setting.Project.PackFormats,
				"task": map[string]interface{}{
					"id":     task.ID,
					"os":     task.OS,
					"arch":   task.Arch,
					"tags":   task.Tags,
					"commit": task.Commit,
				},
			})
		} else {
			ctx.Status(204)
		}
		return
	}

	task, err := models.GetTaskByID(ctx.Builder.TaskID)
	if err != nil {
		ctx.Error("GetTaskByID [%d]: %v", ctx.Builder.TaskID, err)
		return
	}

	switch status {
	case "UPLOADING":
		task.Status = models.TASK_STATUS_UPLOADING
		if err = task.Save(); err != nil {
			ctx.Error("Save: %v", err)
			return
		}
	case "FAILED":
		if err = task.BuildFailed(); err != nil {
			ctx.Error("BuildFailed: %v", err)
			return
		}
	case "SUCCEED":
		if err = task.BuildSucceed(); err != nil {
			ctx.Error("BuildSucceed: %v", err)
			return
		}
	}

	ctx.Status(204)
}

func UploadArtifact(ctx *context.Context) {
	log.Trace("Receiving artifact from builder '%d' for task '%d'", ctx.Builder.ID, ctx.Builder.TaskID)

	task, err := models.GetTaskByID(ctx.Builder.TaskID)
	if err != nil {
		ctx.Error("GetTaskByID: %v", err)
		return
	}

	if err = ctx.Req.ParseMultipartForm(1024 * 1024 * 32); err != nil {
		ctx.Error("ParseMultipartForm: %v", err)
		return
	}

	savePath := path.Join("data/artifacts", task.ArtifactName(ctx.Req.Header.Get("X-LUBAN-FORMAT")))
	os.MkdirAll(path.Dir(savePath), os.ModePerm)

	fw, err := os.Create(savePath)
	if err != nil {
		ctx.Error("Create: %v", err)
		return
	}
	defer fw.Close()

	fr, _, err := ctx.Req.FormFile("artifact")
	if err != nil {
		ctx.Error("FormFile: %v", err)
		return
	}
	defer fr.Close()

	if _, err = io.Copy(fw, fr); err != nil {
		ctx.Error("Copy: %v", err)
		return
	}

	ctx.Status(204)
}
