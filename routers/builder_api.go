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

	"github.com/lubanstudio/luban/models"
	"github.com/lubanstudio/luban/modules/context"
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

	matrices := make([]*models.Matrix, 0, 2)
	if err = json.Unmarshal(data, &matrices); err != nil {
		ctx.Error(500, fmt.Sprintf("json.Unmarshal: %v", err))
		return
	} else if err = ctx.Builder.UpdateMatrices(matrices); err != nil {
		ctx.Error(500, fmt.Sprintf("UpdateMatrices: %v", err))
		return
	}

	ctx.Status(204)
}

func HeartBeat(ctx *context.Context) {
	if err := ctx.Builder.HeartBeat(ctx.Req.Header.Get("X-LUBAN-STATUS") == "IDLE"); err != nil {
		ctx.Error(500, fmt.Sprintf("HeartBeat: %v", err))
		return
	}

	ctx.Status(204)
}
