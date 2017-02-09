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

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	log "gopkg.in/clog.v1"

	"github.com/lubanstudio/luban/modules/setting"
)

var x *gorm.DB

func init() {
	var err error
	x, err = gorm.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=true",
		setting.Database.User, setting.Database.Password, setting.Database.Host, setting.Database.Name))
	if err != nil {
		log.Fatal(4, "Fail to connect database: %s", err)
	}

	if err = x.Set("gorm:table_options", "ENGINE=InnoDB").
		AutoMigrate(new(User), new(Builder), new(Matrix), new(Task)).Error; err != nil {
		log.Fatal(4, "Fail to auto migrate database: %s", err)
	}
}

func releaseTransaction(tx *gorm.DB) {
	if tx.Error != nil {
		tx.Rollback()
	}
}

func IsErrRecordNotFound(err error) bool {
	return err == gorm.ErrRecordNotFound
}

func Count(bean interface{}) int64 {
	var count int64
	x.Model(bean).Count(&count)
	return count
}
