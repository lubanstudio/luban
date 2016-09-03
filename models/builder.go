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
	"strings"
	"time"

	"github.com/Unknwon/com"

	"github.com/lubanstudio/luban/modules/tool"
)

type TrustLevel int

const (
	TRUST_LEVEL_UNAPPROVED TrustLevel = iota
	TRUST_LEVEL_APPROVED
	TRUST_LEVEL_OFFICIAL TrustLevel = 99
)

func (l TrustLevel) ToString() string {
	switch l {
	case TRUST_LEVEL_APPROVED:
		return "Approved"
	case TRUST_LEVEL_OFFICIAL:
		return "Official"
	}
	return "Unapproved"
}

func ParseTrustLevel(n int) TrustLevel {
	switch n {
	case 1:
		return TRUST_LEVEL_APPROVED
	case 99:
		return TRUST_LEVEL_OFFICIAL
	default:
		return TRUST_LEVEL_UNAPPROVED
	}
}

type Builder struct {
	ID         int64
	Name       string
	Token      string `gorm:"UNIQUE"`
	TrustLevel TrustLevel

	IsIdle        bool `gorm:"NOT NULL"`
	LastHeartBeat int64
	Created       int64

	TaskID int64
}

func (b *Builder) BeforeCreate() {
	b.Created = time.Now().Unix()
}

func (b *Builder) Status() string {
	if b.LastHeartBeat < time.Now().Add(-1*time.Minute).Unix() {
		return "Offline"
	}
	if b.IsIdle {
		return "Idle"
	}
	return "Busy"
}

func (b *Builder) CreatedTime() time.Time {
	return time.Unix(b.Created, 0)
}

// HeartBeat updates last active and status.
func (b *Builder) HeartBeat(isIdle bool) error {
	b.LastHeartBeat = time.Now().Unix()
	b.IsIdle = isIdle
	return b.Save()
}

func (b *Builder) Save() error {
	if !IsErrRecordNotFound(x.Where("name = ? AND id != ?", b.Name, b.ID).First(new(Builder)).Error) {
		return ErrBuilderExists{b.Name}
	}
	return x.Save(b).Error
}

func NewBuilder(name string) (*Builder, error) {
	if !IsErrRecordNotFound(x.Where("name = ?", name).First(new(Builder)).Error) {
		return nil, ErrBuilderExists{name}
	}

	builder := &Builder{
		Name:       name,
		Token:      tool.NewSecretToekn(),
		TrustLevel: 1,
	}
	return builder, x.Create(builder).Error
}

func GetBuilderByID(id int64) (*Builder, error) {
	builder := new(Builder)
	return builder, x.First(builder, id).Error
}

func GetBuilderByToken(token string) (*Builder, error) {
	builder := new(Builder)
	return builder, x.Where("token = ?", token).First(builder).Error
}

func ListBuilders() ([]*Builder, error) {
	builders := make([]*Builder, 0, 10)
	return builders, x.Find(&builders).Error
}

func CountBuilders() int64 {
	return Count(new(Builder))
}

func RegenerateBuilderToken(id int64) error {
	return x.First(new(Builder), id).Update("token", tool.NewSecretToekn()).Error
}

// TODO: delete building history and matrices
func DeleteBuilderByID(id int64) error {
	return x.Delete(new(Builder), id).Error
}

func MatchBuilders(os, arch string, tags []string) ([]int64, error) {
	matrices, err := FindMatrices(os, arch)
	if err != nil {
		return nil, fmt.Errorf("FindBuilders: %v", err)
	}
	if len(matrices) == 0 {
		return nil, ErrNoSuitableMatrix{os, arch, tags}
	}

	marked := make(map[int64]bool)
	builderIDs := make([]int64, 0, 5)
CHECK_TAG:
	for _, m := range matrices {
		supportTags := strings.Split(m.Tags, ",")

		for _, tag := range tags {
			if !com.IsSliceContainsStr(supportTags, tag) {
				continue CHECK_TAG
			}
		}

		if !marked[m.BuilderID] {
			marked[m.BuilderID] = true
			builderIDs = append(builderIDs, m.BuilderID)
		}
	}
	return builderIDs, nil
}
