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

package models

import (
	"fmt"
)

type Matrix struct {
	ID        int64
	BuilderID int64
	OS        string
	Arch      string
	Tags      string
}

func UpdateBuilderMatrices(builderID int64, matrices []*Matrix) error {
	if err := x.Delete(new(Matrix), "builder_id = ?", builderID).Error; err != nil {
		return fmt.Errorf("delete old matrices: %v", err)
	}

	tx := x.Begin()
	defer releaseTransaction(tx)

	for _, matrix := range matrices {
		matrix.BuilderID = builderID
		if err := tx.Create(matrix).Error; err != nil {
			return fmt.Errorf("create matrix: %v", err)
		}
	}

	return tx.Commit().Error
}

func (b *Builder) UpdateMatrices(matrices []*Matrix) error {
	return UpdateBuilderMatrices(b.ID, matrices)
}
