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
	"fmt"
	"os"

	"github.com/Unknwon/com"
	log "gopkg.in/clog.v1"
	"gopkg.in/ini.v1"
)

var (
	AppVer   string
	ProdMode bool

	HTTPPort      int
	ArtifactsPath string

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
		Name        string
		CloneURL    string `ini:"CLONE_URL"`
		CommitURL   string `ini:"COMMIT_URL"`
		ImportPath  string
		Branches    []string
		PackRoot    string
		PackEntries []string
		PackFormats []string
	}

	Cfg *ini.File
)

func init() {
	err := log.New(log.CONSOLE, log.ConsoleConfig{})
	if err != nil {
		fmt.Printf("Fail to create new logger: %v\n", err)
		os.Exit(1)
	}

	Cfg, err = ini.Load("conf/app.ini")
	if err != nil {
		log.Fatal(4, "Fail to load configuration: %s", err)
	}
	if com.IsFile("custom/app.ini") {
		if err = Cfg.Append("custom/app.ini"); err != nil {
			log.Fatal(4, "Fail to load custom configuration: %s", err)
		}
	}
	Cfg.NameMapper = ini.AllCapsUnderscore

	ProdMode = Cfg.Section("").Key("RUN_MODE").String() == "prod"
	if ProdMode {
		if err := log.New(log.FILE, log.FileConfig{
			Level:    log.INFO,
			Filename: "log/luban.log",
			FileRotationConfig: log.FileRotationConfig{
				Rotate:  true,
				Daily:   true,
				MaxDays: 3,
			},
		}); err != nil {
			log.Fatal(0, "Fail to create new logger: %v", err)
		}
		log.Delete(log.CONSOLE)
	}

	HTTPPort = Cfg.Section("").Key("HTTP_PORT").MustInt(8086)
	ArtifactsPath = Cfg.Section("").Key("ARTIFACTS_PATH").MustString("data/artifacts")

	if err = Cfg.Section("database").MapTo(&Database); err != nil {
		log.Fatal(4, "Fail to map section 'database': %v", err)
	} else if err = Cfg.Section("oauth2").MapTo(&OAuth2); err != nil {
		log.Fatal(4, "Fail to map section 'oauth2': %v", err)
	} else if err = Cfg.Section("project").MapTo(&Project); err != nil {
		log.Fatal(4, "Fail to map section 'project': %v", err)
	}

	if err = loadMatrices(); err != nil {
		log.Fatal(4, "loadMatrices: %v", err)
	} else if err = loadBatchJobs(); err != nil {
		log.Fatal(4, "loadBatchJobs: %v", err)
	}
}
