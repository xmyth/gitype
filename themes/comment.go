// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package themes

import (
	"strconv"
	"time"
)

type Comment struct {
	post *Post

	ID         int64
	Created    int64
	IP         string
	Agent      string
	Content    string
	IsAdmin    bool
	AuthorName string
	AuthorURL  string
}

func (c *Comment) CreatedFormat() string {
	return time.Unix(c.Created, 0).Format(opt.DateFormat)
}

func (c *Comment) Permalink() string {
	return c.post.Permalink() + "#comments-" + strconv.FormatInt(c.ID, 10)
}
