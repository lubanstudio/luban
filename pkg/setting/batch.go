// Copyright 2017 Unknwon
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

package setting

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sort"
)

var BatchTasks []*BatchTask

type BatchTask struct {
	OS   string   `json:"os"`
	Arch string   `json:"arch"`
	Tags []string `json:"tags"`
}

func loadBatchJobs() error {
	data, err := ioutil.ReadFile("custom/batch.json")
	if err != nil {
		return fmt.Errorf("ReadFile: %v", err)
	}

	if err = json.Unmarshal(data, &BatchTasks); err != nil {
		return fmt.Errorf("Unmarshal: %v", err)
	}

	for i := range BatchTasks {
		sort.Strings(BatchTasks[i].Tags)
	}

	return nil
}
