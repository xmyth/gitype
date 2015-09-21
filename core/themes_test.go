// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package core

import (
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"
)

func TestLoadThemeFile(t *testing.T) {
	a := assert.New(t)

	theme, err := loadThemeFile("./testdata/theme1/theme.json")
	a.NotError(err)
	a.Equal(theme.Name, "default").Equal(theme.Author.Name, "caixw")
}

func TestThemes_Themes(t *testing.T) {
	a := assert.New(t)

	ts, err := LoadThemes("./testdata", "theme1")
	a.NotError(err).NotNil(ts)

	themes := ts.Themes()
	a.Equal(2, len(themes))
	a.Equal(themes["theme1"].Name, "default")
}

func TestThemes_LoadTheme_Render(t *testing.T) {
	a := assert.New(t)
	ts, err := LoadThemes("./testdata", "theme1")
	a.NotError(err).NotNil(ts)

	a.NotError(ts.LoadTheme("theme2"))
	w := httptest.NewRecorder()
	ts.Render(w, "index", nil)
	a.Equal("theme2", w.Body.String())
}