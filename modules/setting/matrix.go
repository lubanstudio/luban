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

package setting

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/Unknwon/com"
)

var (
	Matrices     []*Matrix
	AllowedOSs   []string
	AllowedArchs []string
	AllowedTags  []string
)

type Matrix struct {
	OS    string   `json:"os"`
	Archs []string `json:"archs"`
	Tags  []string `json:"tags"`
}

func LoadMatrices() error {
	data, err := ioutil.ReadFile("custom/matrices.json")
	if err != nil {
		return fmt.Errorf("ReadFile: %v", err)
	}

	if err = json.Unmarshal(data, &Matrices); err != nil {
		return fmt.Errorf("Unmarshal: %v", err)
	}

	for _, m := range Matrices {
		if !com.IsSliceContainsStr(AllowedOSs, m.OS) {
			AllowedOSs = append(AllowedOSs, m.OS)
		}

		for _, arch := range m.Archs {
			if !com.IsSliceContainsStr(AllowedArchs, arch) {
				AllowedArchs = append(AllowedArchs, arch)
			}
		}

		for _, tag := range m.Tags {
			if !com.IsSliceContainsStr(AllowedTags, tag) {
				AllowedTags = append(AllowedTags, tag)
			}
		}
	}

	return nil
}
