// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package data

import (
	"testing"

	"github.com/issue9/assert"
)

func TestFindTheme(t *testing.T) {
	a := assert.New(t)

	conf := &config{Theme: "no exists"}
	theme, err := findTheme(testdataPath, conf)
	a.Error(err).Nil(theme)

	conf.Theme = "t1"
	theme, err = findTheme(testdataPath, conf)
	a.NotError(err).NotNil(theme)
	a.Equal(theme.ID, "t1")
	a.Equal(theme.Name, "name")
	a.Equal(theme.Author.Name, "caixw")
}

func TestLoadTheme(t *testing.T) {
	a := assert.New(t)

	conf := &config{Theme: "t1"}
	theme, err := loadTheme(testdataPath, conf)
	a.NotError(err).NotNil(theme)

	a.Equal(theme.Name, "name")
	a.Equal(theme.Author.Name, "caixw")
}

func TestStripTags(t *testing.T) {
	a := assert.New(t)

	tests := map[string]string{
		"<div>str</div>":        "str",
		"str<br />":             "str",
		"<div><p>str</p></div>": "str",
	}

	for expr, val := range tests {
		a.Equal(stripTags(expr), val, "测试[%v]时出错", expr)
	}
}
