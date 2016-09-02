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

package form

import (
	"reflect"
	"strings"

	"github.com/Unknwon/com"
	"github.com/go-macaron/binding"
)

type Form interface {
	binding.Validator
}

func init() {
	binding.SetNameMapper(com.ToSnakeCase)
}

// AssignForm assign form values back to the template data.
func AssignForm(form interface{}, data map[string]interface{}) {
	typ := reflect.TypeOf(form)
	val := reflect.ValueOf(form)

	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
		val = val.Elem()
	}

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)

		fieldName := field.Tag.Get("form")
		// Allow ignored fields in the struct
		if fieldName == "-" {
			continue
		} else if len(fieldName) == 0 {
			fieldName = com.ToSnakeCase(field.Name)
		}

		data[fieldName] = val.Field(i).Interface()
	}
}

func getRuleBody(field reflect.StructField, prefix string) string {
	for _, rule := range strings.Split(field.Tag.Get("binding"), ";") {
		if strings.HasPrefix(rule, prefix) {
			return rule[len(prefix) : len(rule)-1]
		}
	}
	return ""
}

func GetSize(field reflect.StructField) string {
	return getRuleBody(field, "Size(")
}

func GetMinSize(field reflect.StructField) string {
	return getRuleBody(field, "MinSize(")
}

func GetMaxSize(field reflect.StructField) string {
	return getRuleBody(field, "MaxSize(")
}

func GetInclude(field reflect.StructField) string {
	return getRuleBody(field, "Include(")
}

func Validate(errs binding.Errors, data map[string]interface{}, f Form) binding.Errors {
	if errs.Len() == 0 {
		return errs
	}

	data["HasError"] = true
	AssignForm(f, data)

	typ := reflect.TypeOf(f)
	val := reflect.ValueOf(f)

	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
		val = val.Elem()
	}

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)

		fieldName := field.Tag.Get("form")
		// Allow ignored fields in the struct
		if fieldName == "-" {
			continue
		}

		if errs[0].FieldNames[0] == field.Name {
			data["Err_"+field.Name] = true

			name := field.Tag.Get("name")
			if len(name) == 0 {
				name = field.Name
			}

			switch errs[0].Classification {
			case binding.ERR_REQUIRED:
				data["ErrorMsg"] = name + " cannot be empty."
			case binding.ERR_ALPHA_DASH:
				data["ErrorMsg"] = name + " must be valid alpha or numeric or dash(-_) characters."
			case binding.ERR_ALPHA_DASH_DOT:
				data["ErrorMsg"] = name + " must be valid alpha or numeric or dash(-_) or dot characters."
			case binding.ERR_SIZE:
				data["ErrorMsg"] = name + " must be size " + GetSize(field)
			case binding.ERR_MIN_SIZE:
				data["ErrorMsg"] = name + " must contain at least " + GetMinSize(field) + " characters."
			case binding.ERR_MAX_SIZE:
				data["ErrorMsg"] = name + " must contain at most " + GetMaxSize(field) + " characters."
			case binding.ERR_EMAIL:
				data["ErrorMsg"] = name + " is not a valid email address."
			case binding.ERR_URL:
				data["ErrorMsg"] = name + " is not a valid URL."
			case binding.ERR_INCLUDE:
				data["ErrorMsg"] = name + " must contain substring '" + GetInclude(field) + "'."
			default:
				data["ErrorMsg"] = "Unknown error: " + errs[0].Classification
			}
			return errs
		}
	}
	return errs
}
