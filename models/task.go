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
	"os"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/Unknwon/com"
	log "gopkg.in/clog.v1"

	"github.com/lubanstudio/luban/pkg/setting"
	"github.com/lubanstudio/luban/pkg/tool"
)

type TaskStatus int

const (
	TASK_STATUS_PENDING TaskStatus = iota
	TASK_STATUS_BUILDING
	TASK_STATUS_UPLOADING
	TASK_STATUS_FAILED
	TASK_STATUS_SUCCEED
	TASK_STATUS_ARCHIVED TaskStatus = 99
)

func (s TaskStatus) ToString() string {
	switch s {
	case TASK_STATUS_PENDING:
		return "Pending"
	case TASK_STATUS_BUILDING:
		return "Building"
	case TASK_STATUS_UPLOADING:
		return "Uploading"
	case TASK_STATUS_FAILED:
		return "Failed"
	case TASK_STATUS_SUCCEED:
		return "Succeed"
	case TASK_STATUS_ARCHIVED:
		return "Archived"
	}
	return "Unapproved"
}

type Task struct {
	ID     int64
	OS     string
	Arch   string
	Tags   string
	Commit string
	Status TaskStatus

	PosterID  int64
	Poster    *User `gorm:"-"`
	BuilderID int64
	Builder   *Builder `gorm:"-"`
	Updated   int64
	Created   int64
}

func (t *Task) BeforeCreate() {
	t.Created = time.Now().Unix()
}

func (t *Task) AfterFind() (err error) {
	if t.PosterID > 0 {
		t.Poster, err = GetUserByID(t.PosterID)
		if err != nil {
			return fmt.Errorf("GetUserByID [%d]: %v", t.PosterID, err)
		}
	}

	if t.BuilderID > 0 {
		t.Builder, err = GetBuilderByID(t.BuilderID)
		if err != nil {
			return fmt.Errorf("GetBuilderByID [%d]: %v", t.BuilderID, err)
		}
	}
	return nil
}

func (t *Task) UpdatedTime() time.Time {
	return time.Unix(t.Updated, 0)
}

func (t *Task) CreatedTime() time.Time {
	return time.Unix(t.Created, 0)
}

func (t *Task) CommitURL() string {
	return com.Expand(setting.Project.CommitURL, map[string]string{"sha": t.Commit})
}

func (t *Task) ArtifactName(format string) string {
	name := setting.Project.PackRoot + "_" + t.Commit[:10] + "_" + t.OS + "_" + t.Arch
	if len(t.Tags) > 0 {
		name += "_" + strings.Replace(t.Tags, ",", "_", -1)
	}
	return name + "." + format
}

func (t *Task) Save() error {
	return x.Save(t).Error
}

func (t *Task) AssignBuilder(builderID int64) (err error) {
	tx := x.Begin()
	defer releaseTransaction(tx)

	t.BuilderID = builderID
	t.Status = TASK_STATUS_BUILDING
	t.Updated = time.Now().Unix()
	if err = tx.Exec("UPDATE builders SET is_idle = ?,task_id = ? WHERE id = ?", false, t.ID, builderID).Error; err != nil {
		return fmt.Errorf("set builder to busy: %v", err)
	} else if err = tx.Save(t).Error; err != nil {
		return fmt.Errorf("save task: %v", err)
	}

	return tx.Commit().Error
}

func (t *Task) buildFinish(status TaskStatus) error {
	builder, err := GetBuilderByID(t.BuilderID)
	if err != nil {
		return fmt.Errorf("GetBuilderByID: %v", err)
	}

	tx := x.Begin()
	defer releaseTransaction(tx)

	t.Status = status
	t.Updated = time.Now().Unix()
	if err = tx.Save(t).Error; err != nil {
		return fmt.Errorf("Save.(task): %v", err)
	}

	builder.IsIdle = true
	builder.TaskID = 0
	if err = tx.Save(builder).Error; err != nil {
		return fmt.Errorf("Save.(builder): %v", err)
	}

	return tx.Commit().Error
}

func (t *Task) BuildFailed() error {
	return t.buildFinish(TASK_STATUS_FAILED)
}

func (t *Task) BuildSucceed() error {
	return t.buildFinish(TASK_STATUS_SUCCEED)
}

func (t *Task) Archive() error {
	t.Status = TASK_STATUS_ARCHIVED
	if err := t.Save(); err != nil {
		return err
	}

	for _, format := range setting.Project.PackFormats {
		os.Remove(path.Join(setting.ArtifactsPath, t.ArtifactName(format)))
	}
	return nil
}

func GetCommitOfBranch(branch string) (string, error) {
	fmt.Println("git", "ls-remote", setting.Project.CloneURL, branch)
	// Get latest commit ID on given branch.
	stdout, stderr, err := com.ExecCmd("git", "ls-remote", setting.Project.CloneURL, branch)
	if err != nil {
		return "", fmt.Errorf("get latest commit of branch '%s': %v - %s", branch, err, stderr)
	}
	if len(stdout) < 40 {
		return "", fmt.Errorf("not enough length of commit ID: %s", stdout)
	}
	return stdout[:40], nil
}

func NewTask(doerID int64, os, arch string, tags []string, branch string) (*Task, error) {
	sort.Strings(tags)

	// Make sure there is a matrix can take the job.
	builderIDs, err := MatchBuilders(os, arch, tags)
	if err != nil {
		if IsErrNoSuitableMatrix(err) {
			return nil, err
		}
		return nil, fmt.Errorf("MatchBuilders: %v", err)
	}
	if len(builderIDs) == 0 {
		return nil, ErrNoSuitableMatrix{os, arch, tags}
	}

	commit, err := GetCommitOfBranch(branch)
	if err != nil {
		return nil, fmt.Errorf("GetCommitOfBranch: %v", err)
	}

	// Check to prevent duplicated tasks
	task := new(Task)
	if err = x.Where("os=? AND arch=? AND tags=? AND commit=? AND status!=? AND status!=?",
		os, arch, strings.Join(tags, ","), commit, TASK_STATUS_FAILED, TASK_STATUS_ARCHIVED).First(task).Error; err == nil {
		return task, nil
	} else if !IsErrRecordNotFound(err) {
		return nil, fmt.Errorf("check existing task: %v", err)
	}

	task = &Task{
		OS:       os,
		Arch:     arch,
		Tags:     strings.Join(tags, ","),
		Commit:   commit,
		PosterID: doerID,
	}
	return task, x.Create(task).Error
}

func NewBatchTasks(doerID int64, branch string) error {
	commit, err := GetCommitOfBranch(branch)
	if err != nil {
		return fmt.Errorf("GetCommitOfBranch: %v", err)
	}

	// Check to prevent duplicated tasks
	for _, t := range setting.BatchTasks {
		task := new(Task)
		if err = x.Where("os=? AND arch=? AND tags=? AND commit=? AND status!=? AND status!=?",
			t.OS, t.Arch, strings.Join(t.Tags, ","), commit, TASK_STATUS_FAILED, TASK_STATUS_ARCHIVED).First(task).Error; err != nil {
			if !IsErrRecordNotFound(err) {
				return fmt.Errorf("check existing task: %v", err)
			}
		}

		task = &Task{
			OS:       t.OS,
			Arch:     t.Arch,
			Tags:     strings.Join(t.Tags, ","),
			Commit:   commit,
			PosterID: doerID,
		}
		if err = x.Create(task).Error; err != nil {
			return fmt.Errorf("create new task: %v", err)
		}
	}

	return nil
}

func GetTaskByID(id int64) (*Task, error) {
	task := new(Task)
	return task, x.First(task, id).Error
}

func ListTasks(page, pageSize int64) ([]*Task, error) {
	tasks := make([]*Task, 0, 10)
	return tasks, x.Limit(pageSize).Offset((page - 1) * pageSize).Order("id DESC").Find(&tasks).Error
}

func ListPendingTasks() ([]*Task, error) {
	tasks := make([]*Task, 0, 10)
	return tasks, x.Where("status = ?", TASK_STATUS_PENDING).Find(&tasks).Error
}

func CountTasks() int64 {
	return Count(new(Task))
}

func AssignTasks() {
	defer func() {
		log.Trace("Finish assigning tasks.")
		time.AfterFunc(30*time.Second, AssignTasks)
	}()

	log.Trace("Start assigning tasks...")
	tasks, err := ListPendingTasks()
	if err != nil {
		log.Error(4, "ListPendingTasks: %v", err)
		return
	}

	for _, t := range tasks {
		var tags []string
		if len(t.Tags) > 0 {
			tags = strings.Split(t.Tags, ",")
		}
		builderIDs, err := MatchBuilders(t.OS, t.Arch, tags)
		if err != nil {
			if !IsErrNoSuitableMatrix(err) {
				log.Error(4, "MatchBuilders [task_id: %d]: %v", t.ID, err)
			}
			continue
		}

		builder := new(Builder)
		if err = x.Where("is_idle = ? AND id IN (?)", true, tool.Int64sToStrings(builderIDs)).First(builder).Error; err != nil {
			if !IsErrRecordNotFound(err) {
				log.Error(4, "find idle builder [task_id: %s]: %v", t.ID, err)
			}
			continue
		}

		if err = t.AssignBuilder(builder.ID); err != nil {
			log.Error(4, "AssignBuilder [task_id: %s, builder_id: %d]: %v", t.ID, builder.ID, err)
			continue
		}

		log.Trace("Assigned task '%d' to builder '%d'", t.ID, builder.ID)
	}
}
