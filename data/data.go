// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package data 负责加载 data 目录下的数据。
// 会调用 github.com/issue9/logs 包的内容，调用之前需要初始化该包。
package data

import "html/template"
import "path/filepath"

// Data 结构体包含了数据目录下所有需要加载的数据内容。
type Data struct {
	Root     string             // Data 数据所在的根目录
	Config   *Config            // 配置内容
	Tags     []*Tag             // map 对顺序是未定的，所以使用 slice
	Links    []*Link            // 友情链接
	Posts    []*Post            // 所有的文章列表
	Themes   map[string]*Theme  // 主题
	Template *template.Template // 当前主题模板，TODO 移到其它地方，当前包只负责加载数据？
}

// Load 函数用于加载一份新的数据。
// root 表示数据在的根目录。
func Load(root string) (*Data, error) {
	d := &Data{
		Root: root,
	}

	if err := d.loadMeta(root); err != nil {
		return nil, err
	}

	// 加载文章
	if err := d.loadPosts(filepath.Join(d.Root, "posts")); err != nil {
		return nil, err
	}

	return d, nil
}

func (d *Data) loadMeta(path string) error {
	// tags
	if err := d.loadTags(d.metaPath("tags.yaml")); err != nil {
		return err
	}

	// links
	if err := d.loadLinks(d.metaPath("links.yaml")); err != nil {
		return err
	}

	// config
	if err := d.loadConfig(d.metaPath("config.yaml")); err != nil {
		return err
	}

	// theme
	themes, err := getThemesName(filepath.Join(d.Root, "themes", d.Config.Theme))
	if err != nil {
		return err
	}
	found := false
	for _, theme := range themes {
		if theme == d.Config.Theme {
			found = true
			break
		}
	}
	if !found {
		return &FieldError{File: "config.yaml", Message: "该主题并不存在", Field: "Theme"}
	}

	// 加载主题的模板
	return d.loadTemplate(filepath.Join(d.Root, "themes"))
}

func (d *Data) metaPath(file string) string {
	return filepath.Join(d.Root, "meta", file)
}
