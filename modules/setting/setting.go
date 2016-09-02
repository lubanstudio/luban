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
	log "github.com/Sirupsen/logrus"
	"github.com/Unknwon/com"
	"gopkg.in/ini.v1"
)

var (
	AppVer  string
	RunMode string

	HTTPPort int

	Database struct {
		Host     string
		Name     string
		User     string
		Password string
	}

	OAuth2 struct {
		ClientID     string `ini:"CLIENT_ID"`
		ClientSecret string
	}

	Project struct {
		Name     string
		CloneURL string `ini:"CLONE_URL"`
		Branches []string
	}

	Cfg *ini.File
)

func init() {
	var err error
	Cfg, err = ini.Load("conf/app.ini")
	if err != nil {
		log.Fatalf("Fail to load configuration: %s", err)
	}
	if com.IsFile("custom/app.ini") {
		if err = Cfg.Append("custom/app.ini"); err != nil {
			log.Fatalf("Fail to load custom configuration: %s", err)
		}
	}
	Cfg.NameMapper = ini.AllCapsUnderscore

	RunMode = Cfg.Section("").Key("RUN_MODE").String()
	if RunMode != "prod" {
		log.SetLevel(log.DebugLevel)
	}

	HTTPPort = Cfg.Section("").Key("HTTP_PORT").MustInt(8086)

	if err = Cfg.Section("database").MapTo(&Database); err != nil {
		log.Fatalf("Fail to map section 'database': %s", err)
	} else if err = Cfg.Section("oauth2").MapTo(&OAuth2); err != nil {
		log.Fatalf("Fail to map section 'oauth2': %s", err)
	} else if err = Cfg.Section("project").MapTo(&Project); err != nil {
		log.Fatalf("Fail to map section 'project': %s", err)
	}

	if err = LoadMatrices(); err != nil {
		log.Fatalf("LoadMatrices: %s", err)
	}
}
