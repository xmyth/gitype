// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package client

import (
	"net/http"
	"strconv"

	"github.com/caixw/gitype/data"
	"github.com/caixw/gitype/vars"
	"github.com/issue9/logs"
	"github.com/issue9/middleware/compress"
	"github.com/issue9/mux"
)

func (client *Client) initRoutes() (err error) {
	handle := func(pattern string, h http.HandlerFunc) {
		if err != nil {
			return
		}

		client.patterns = append(client.patterns, pattern)
		err = client.mux.HandleFunc(pattern, client.prepare(h), http.MethodGet)
	}

	handle(vars.PostURL("{slug}"), client.getPost)   // posts/2016/about.html   posts/{slug}.html
	handle(vars.AssetURL("{path}"), client.getAsset) // posts/2016/about/abc.png  posts/{path}
	handle(vars.IndexURL(0), client.getPosts)        // index.html
	handle(vars.LinksURL(), client.getLinks)         // links.html
	handle(vars.TagURL("{slug}", 1), client.getTag)  // tags/tag1.html     tags/{slug}.html
	handle(vars.TagsURL(), client.getTags)           // tags.html
	handle(vars.ArchivesURL(), client.getArchives)   // archives.html
	handle(vars.SearchURL("", 1), client.getSearch)  // search.html
	handle(vars.ThemeURL("{path}"), client.getTheme) // themes/...          themes/{path}
	handle("/{path}", client.getRaw)                 // /...                /{path}

	return err
}

// 文章详细页
// /posts/{slug}.html
func (client *Client) getPost(w http.ResponseWriter, r *http.Request) {
	slug, err := mux.Params(r).String("slug")
	if err != nil {
		logs.Error(err)
		client.getAsset(w, r)
		return
	}

	index := -1
	for i, p := range client.data.Posts {
		if p.Slug == slug {
			index = i
			break
		}
	}

	if index < 0 {
		logs.Debugf("并未找到与之相对应的文章：%s", slug)
		client.getRaw(w, r) // 文章不存在，则查找 raws 目录下是否存在同名文件
		return
	}

	post := client.data.Posts[index]
	p := client.page(vars.PagePost, w, r)

	p.Post = post
	p.Keywords = post.Keywords
	p.Description = post.Summary
	p.Title = post.HTMLTitle
	p.Canonical = client.data.BuildURL(post.Permalink)
	p.License = post.License // 文章可具体指定协议
	p.Author = post.Author   // 文章可具体指定作者

	if index > 0 {
		prev := client.data.Posts[index-1]
		p.prevPage(prev.Permalink, prev.Title)
	}
	if index+1 < len(client.data.Posts) {
		next := client.data.Posts[index+1]
		p.nextPage(next.Permalink, next.Title)
	}

	p.render(post.Template)
}

// 首页及文章列表页
// /
// /index.html?page=2
func (client *Client) getPosts(w http.ResponseWriter, r *http.Request) {
	page, ok := client.queryInt(w, r, vars.URLQueryPage, 1)
	if !ok {
		return
	}

	if page < 1 {
		logs.Debugf("请求的页码[%d]小于1\n", page)
		client.renderError(w, r, http.StatusNotFound) // 页码为负数的表示不存在，跳转到 404 页面
		return
	}

	p := client.page(vars.PageIndex, w, r)
	if page > 1 { // 非首页，标题显示页码数
		p.Type = vars.PagePosts
	}
	pp := client.data.Pages[vars.PagePosts]
	p.Title = pp.Title
	p.Keywords = pp.Keywords
	p.Description = pp.Description
	p.Canonical = client.data.BuildURL(vars.PostsURL(page))

	start, end, ok := client.getPostsRange(len(client.data.Posts), page, w, r)
	if !ok {
		return
	}
	p.Posts = client.data.Posts[start:end]
	if page > 1 {
		p.prevPage(vars.PostsURL(page-1), "")
	}
	if end < len(client.data.Posts) {
		p.nextPage(vars.PostsURL(page+1), "")
	}

	p.render(vars.PagePosts)
}

// 标签详细页
// /tags/tag1.html?page=2
func (client *Client) getTag(w http.ResponseWriter, r *http.Request) {
	slug, err := mux.Params(r).String("slug")
	if err != nil {
		logs.Error(err)
		client.getRaw(w, r)
		return
	}

	var tag *data.Tag
	for _, t := range client.data.Tags {
		if t.Slug == slug {
			tag = t
			break
		}
	}

	if tag == nil {
		logs.Debugf("查找的标签 %s 不存在", slug)
		client.getRaw(w, r) // 标签不存在，则查找该文件是否存在于 raws 目录下。
		return
	}

	page, ok := client.queryInt(w, r, vars.URLQueryPage, 1)
	if !ok {
		return
	}
	if page < 1 {
		logs.Debugf("请求的页码[%d]小于 1", page)
		client.renderError(w, r, http.StatusNotFound) // 页码为负数的表示不存在，跳转到 404 页面
		return
	}

	p := client.page(vars.PageTag, w, r)
	p.Tag = tag
	p.Title = tag.HTMLTitle
	p.Keywords = tag.Keywords
	p.Description = tag.Content
	p.Canonical = client.data.BuildURL(vars.TagURL(slug, page))

	start, end, ok := client.getPostsRange(len(tag.Posts), page, w, r)
	if !ok {
		return
	}
	p.Posts = tag.Posts[start:end]
	if page > 1 {
		p.prevPage(vars.TagURL(slug, page-1), "")
	}
	if end < len(tag.Posts) {
		p.nextPage(vars.TagURL(slug, page+1), "")
	}

	p.render(vars.PageTag)
}

// 友情链接页
// /links.html
func (client *Client) getLinks(w http.ResponseWriter, r *http.Request) {
	p := client.page(vars.PageLinks, w, r)
	pp := client.data.Pages[vars.PageLinks]
	p.Title = pp.Title
	p.Keywords = pp.Keywords
	p.Description = pp.Description
	p.Canonical = client.data.BuildURL(vars.LinksURL())

	p.render(vars.PageLinks)
}

// 标签列表页
// /tags.html
func (client *Client) getTags(w http.ResponseWriter, r *http.Request) {
	p := client.page(vars.PageTags, w, r)
	pp := client.data.Pages[vars.PageTags]
	p.Title = pp.Title
	p.Keywords = pp.Keywords
	p.Description = pp.Description
	p.Canonical = client.data.BuildURL(vars.TagsURL())

	p.render(vars.PageTags)
}

// 归档页
// /archives.html
func (client *Client) getArchives(w http.ResponseWriter, r *http.Request) {
	p := client.page(vars.PageArchives, w, r)
	pp := client.data.Pages[vars.PageArchives]
	p.Title = pp.Title
	p.Keywords = pp.Keywords
	p.Description = pp.Description
	p.Canonical = client.data.BuildURL(vars.ArchivesURL())
	p.Archives = client.data.Archives

	p.render(vars.PageArchives)
}

// 确认当前文章列表页选择范围。
func (client *Client) getPostsRange(postsSize, page int, w http.ResponseWriter, r *http.Request) (start, end int, ok bool) {
	size := client.data.PageSize
	start = size * (page - 1) // 系统从零开始计数
	if start > postsSize {
		logs.Debugf("请求页码为[%d]，实际文章数量为[%d]\n", page, postsSize)
		client.renderError(w, r, http.StatusNotFound) // 页码超出范围，不存在
		return 0, 0, false
	}

	end = start + size
	if postsSize < end {
		end = postsSize
	}

	return start, end, true
}

// 每次访问前需要做的预处理工作。
func (client *Client) prepare(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logs.Infof("%s: %s", r.UserAgent(), r.URL) // 输出访问日志

		// 直接根据整个博客的最后更新时间来确认 etag
		if r.Header.Get("If-None-Match") == client.etag {
			logs.Infof("304: %s", r.URL)
			w.WriteHeader(http.StatusNotModified)
			return
		}
		w.Header().Set("Etag", client.etag)
		w.Header().Set("Content-Language", client.data.Language)
		compress.New(f, logs.ERROR()).ServeHTTP(w, r)
	}
}

// 获取查询参数 key 的值，并将其转换成 Int 类型，若该值不存在返回 def 作为其默认值，
// 若是类型不正确，则返回一个 false，并向客户端输出一个 400 错误。
func (client *Client) queryInt(w http.ResponseWriter, r *http.Request, key string, def int) (int, bool) {
	val := r.FormValue(key)
	if len(val) == 0 {
		return def, true
	}

	ret, err := strconv.Atoi(val)
	if err != nil {
		logs.Error(err)
		client.renderError(w, r, http.StatusBadRequest)
		return 0, false
	}
	return ret, true
}
