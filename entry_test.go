package main

import (
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEntryFromReader(t *testing.T) {
	f, err := os.Open("example/data/karimen.hatenablog.com/entry/2012/12/18/000000.md")
	assert.NoError(t, err)

	jst, err := time.LoadLocation("Asia/Tokyo")
	assert.NoError(t, err)

	e, err := entryFromReader(f)
	assert.NoError(t, err)

	assert.Equal(t, e.Title, "所内 #2")

	assert.True(t, e.Date.Equal(time.Date(2012, 12, 18, 0, 0, 0, 0, jst)))
}

var content = `---
Title:   所内 #3
Date:    2012-12-19T00:00:00+09:00
URL:     http://hatenablog.example.com/1
EditURL: http://hatenablog.example.com/1/edit
---
test
test2
`

func TestFullContent(t *testing.T) {
	u, _ := url.Parse("http://hatenablog.example.com/1")
	jst, _ := time.LoadLocation("Asia/Tokyo")
	d := time.Date(2012, 12, 19, 0, 0, 0, 0, jst)

	e := &Entry{
		URL:          u,
		EditURL:      u.String() + "/edit",
		Title:        "所内 #3",
		Date:         &d,
		LastModified: &d,
		Content:      "test\ntest2",
	}
	assert.Equal(t, e.fullContent(), content)
}

func TestFrontmatterEntryFromReader(t *testing.T) {
	jst, _ := time.LoadLocation("Asia/Tokyo")

	e, err := entryFromReader(strings.NewReader(content))
	assert.NoError(t, err)

	assert.Equal(t, e.Title, "所内 #3")
	assert.True(t, e.Date.Equal(time.Date(2012, 12, 19, 0, 0, 0, 0, jst)))
	assert.Equal(t, e.URL.String(), "http://hatenablog.example.com/1")
	assert.Equal(t, e.EditURL, "http://hatenablog.example.com/1/edit")
	assert.Equal(t, e.Content, "test\ntest2\n")
}

var draftContent = `---
Title:   所内 #4
Date:    2012-12-20T00:00:00+09:00
URL:     http://hatenablog.example.com/2
EditURL: http://hatenablog.example.com/2/edit
Draft:   yes
---
下書き
`

func TestDraftFullContent(t *testing.T) {
	u, _ := url.Parse("http://hatenablog.example.com/2")
	jst, _ := time.LoadLocation("Asia/Tokyo")
	d := time.Date(2012, 12, 20, 0, 0, 0, 0, jst)

	e := &Entry{
		URL:          u,
		EditURL:      u.String() + "/edit",
		Title:        "所内 #4",
		Date:         &d,
		LastModified: &d,
		IsDraft:      true,
		Content:      "下書き\n",
	}
	assert.Equal(t, e.fullContent(), draftContent)
}

func TestFrontmatterDraftEntryFromReader(t *testing.T) {
	jst, _ := time.LoadLocation("Asia/Tokyo")

	e, err := entryFromReader(strings.NewReader(draftContent))
	assert.NoError(t, err)

	assert.Equal(t, e.Title, "所内 #4")
	assert.True(t, e.Date.Equal(time.Date(2012, 12, 20, 0, 0, 0, 0, jst)))
	assert.Equal(t, e.URL.String(), "http://hatenablog.example.com/2")
	assert.Equal(t, e.EditURL, "http://hatenablog.example.com/2/edit")
	assert.True(t, e.IsDraft)
	assert.Equal(t, e.Content, "下書き\n")
}
