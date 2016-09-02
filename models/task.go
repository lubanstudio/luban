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
	"sort"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/Unknwon/com"

	"github.com/lubanstudio/luban/modules/setting"
	"github.com/lubanstudio/luban/modules/tool"
)

type TaskStatus int

const (
	TASK_STATUS_PENDING TaskStatus = iota
	TASK_STATUS_RUNNING
	TASK_STATUS_UPLOADING
	TASK_STATUS_FAILED
	TASK_STATUS_SUCCEED
)

func (s TaskStatus) ToString() string {
	switch s {
	case TASK_STATUS_PENDING:
		return "Pending"
	case TASK_STATUS_RUNNING:
		return "Running"
	case TASK_STATUS_UPLOADING:
		return "Uploading"
	case TASK_STATUS_FAILED:
		return "Failed"
	case TASK_STATUS_SUCCEED:
		return "Succeed"
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
	BuilderID int64
	Updated   int64
	Created   int64
}

func (t *Task) BeforeCreate() {
	t.Created = time.Now().Unix()
}

func (t *Task) UpdatedTime() time.Time {
	return time.Unix(t.Updated, 0)
}

func (t *Task) CreatedTime() time.Time {
	return time.Unix(t.Created, 0)
}

func (t *Task) AssignBuilder(builderID int64) (err error) {
	tx := x.Begin()
	defer releaseTransaction(tx)

	t.BuilderID = builderID
	t.Status = TASK_STATUS_RUNNING
	if err = tx.Exec("UPDATE builders SET is_idle = ? AND task_id = ? WHERE id = ?", false, t.ID, builderID).Error; err != nil {
		return fmt.Errorf("set builder to busy: %v", err)
	} else if err = tx.Save(t).Error; err != nil {
		return fmt.Errorf("save task: %v", err)
	}

	return tx.Commit().Error
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

	// Get latest commit ID on given branch.
	stdout, stderr, err := com.ExecCmd("git", "ls-remote", setting.Project.CloneURL, branch)
	if err != nil {
		return nil, fmt.Errorf("get latest commit of branch '%s': %v - %s", branch, err, stderr)
	}
	if len(stdout) < 40 {
		return nil, fmt.Errorf("not enough length of commit ID: %s", stdout)
	}
	commit := stdout[:40]

	// Check to prevent duplicated tasks.
	task := new(Task)
	if err = x.Where("os = ? AND arch = ? AND tags = ? AND commit = ?",
		os, arch, strings.Join(tags, ","), commit).First(task).Error; err == nil {
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

func GetTaskByID(id int64) (*Task, error) {
	task := new(Task)
	return task, x.First(task, id).Error
}

func ListTasks() ([]*Task, error) {
	tasks := make([]*Task, 0, 10)
	return tasks, x.Find(&tasks).Error
}

func ListPendingTasks() ([]*Task, error) {
	tasks := make([]*Task, 0, 10)
	return tasks, x.Where("status = ?", TASK_STATUS_PENDING).Find(&tasks).Error
}

func AssignTasks() {
	defer func() {
		log.Debugln("Finish assigning tasks.")
		time.AfterFunc(60*time.Second, AssignTasks)
	}()

	log.Debugln("Start assigning tasks...")
	tasks, err := ListPendingTasks()
	if err != nil {
		log.Errorf("ListPendingTasks: %v", err)
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
				log.Errorf("MatchBuilders [task_id: %d]: %v", t.ID, err)
			}
			continue
		}

		builder := new(Builder)
		if err = x.Where("is_idle = ? AND id IN (?)", true, tool.Int64sToStrings(builderIDs)).First(builder).Error; err != nil {
			if !IsErrRecordNotFound(err) {
				log.Errorf("find idle builder [task_id: %s]: %v", t.ID, err)
			}
			continue
		}

		if err = t.AssignBuilder(builder.ID); err != nil {
			log.Errorf("AssignBuilder [task_id: %s, builder_id: %d]: %v", t.ID, builder.ID, err)
			continue
		}
	}
}
