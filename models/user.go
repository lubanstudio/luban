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
	"encoding/json"
	"fmt"
	"time"

	"github.com/Unknwon/com"
	"github.com/parnurzeal/gorequest"
)

type User struct {
	ID        int64
	OAuthID   string `gorm:"column:oauth_id;UNIQUE"`
	GitHubID  string `gorm:"column:github_id;UNIQUE"`
	Username  string
	AvatarURL string
	IsAdmin   bool `gorm:"NOT NULL"`
	Created   int64
}

func (u *User) BeforeCreate() {
	u.Created = time.Now().Unix()
}

func GetUserByID(id int64) (*User, error) {
	user := new(User)
	return user, x.Where("id = ?", id).First(user).Error
}

func GetUserByGitHubID(githubID string) (*User, error) {
	user := new(User)
	return user, x.Where("github_id = ?", githubID).First(user).Error
}

func GetUserByOAuthID(oauthID string) (*User, error) {
	user := new(User)
	return user, x.Where("oauth_id = ?", oauthID).First(user).Error
}

// GetOrCreateUserByGitHubID retrieves a user based on GitHub ID,
// and creates a new user if does not exists.
// It returns true if a new user created.
func GetOrCreateUserByGitHubID(oauthID, githubID, username, avatarURL string) (*User, bool, error) {
	user, err := GetUserByGitHubID(githubID)
	if err != nil && !IsErrRecordNotFound(err) {
		return nil, false, fmt.Errorf("GetUserByGitHubID: %v", err)
	}

	isNew := false
	if IsErrRecordNotFound(err) {
		user.OAuthID = oauthID
		user.GitHubID = githubID
		user.Username = username
		user.AvatarURL = avatarURL
		if err = x.Create(user).Error; err != nil {
			return nil, false, fmt.Errorf("create new user: %v", err)
		}
		isNew = true
	}

	// Update OAuthID as needed
	if len(oauthID) > 0 && user.OAuthID != oauthID {
		user.OAuthID = oauthID
		user.AvatarURL = avatarURL
		if err = x.Save(user).Error; err != nil {
			return nil, false, fmt.Errorf("update user OAuthID: %v", err)
		}
	}

	// Make the first user be admin
	if Count(new(User)) == 1 {
		user.IsAdmin = true
		if err = x.Save(user).Error; err != nil {
			return nil, false, fmt.Errorf("set user as admin: %v", err)
		}
	}

	return user, isNew, nil
}

// GetOrCreateUserByOAuthID retrieves a user based on OAuth ID,
// and creates a new user if does not exists.
// It returns true if a new user created.
func GetOrCreateUserByOAuthID(oauthID string) (*User, bool, error) {
	user, err := GetUserByOAuthID(oauthID)
	if err != nil && !IsErrRecordNotFound(err) {
		return nil, false, fmt.Errorf("GetUserByOAuthID: %v", err)
	}

	isNew := false
	if IsErrRecordNotFound(err) {
		// Fetch user info
		_, data, errs := gorequest.New().Get("https://api.github.com/user").Query("access_token=" + oauthID).EndBytes()
		if len(errs) > 0 {
			return nil, false, fmt.Errorf("request GitHub user info: %v", errs[0])
		}
		infos := make(map[string]interface{})
		if err = json.Unmarshal(data, &infos); err != nil {
			return nil, false, fmt.Errorf("decoding GitHub user info: %v", err)
		}
		if infos["id"] == nil || infos["login"] == nil {
			return nil, false, fmt.Errorf("'id' or 'login' not found in returned GitHub user info: %v", err)
		}
		user, isNew, err = GetOrCreateUserByGitHubID(oauthID,
			com.ToStr(infos["id"]), com.ToStr(infos["login"]), com.ToStr(infos["avatar_url"]))
		if err != nil {
			return nil, false, fmt.Errorf("GetOrCreateUserByGitHubID: %v", err)
		}
	}

	return user, isNew, nil
}
