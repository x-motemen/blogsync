package main

import (
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

func TestEntryFromReader(t *testing.T) {
	f, err := os.Open("example/data/karimen.hatenablog.com/entry/2012/12/18/000000.md")
	assert.NoError(t, err)
	defer f.Close()

	jst, err := time.LoadLocation("Asia/Tokyo")
	assert.NoError(t, err)

	e, err := entryFromReader(f)
	assert.NoError(t, err)

	assert.Equal(t, e.Title, "所内#2")

	assert.True(t, e.Date.Equal(time.Date(2012, 12, 18, 0, 0, 0, 0, jst)))
}

var content = `---
Title: 所内#3
Date: 2012-12-19T00:00:00+09:00
URL: http://hatenablog.example.com/1
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
		EntryHeader: &EntryHeader{
			URL:     &EntryURL{u},
			EditURL: u.String() + "/edit",
			Title:   "所内#3",
			Date:    &entryTime{&d},
		},
		LastModified: &d,
		Content:      "test\ntest2",
	}
	assert.Equal(t, e.fullContent(), content)
}

func TestFrontmatterEntryFromReader(t *testing.T) {
	jst, _ := time.LoadLocation("Asia/Tokyo")

	e, err := entryFromReader(strings.NewReader(content))
	assert.NoError(t, err)

	assert.Equal(t, e.Title, "所内#3")
	assert.True(t, e.Date.Equal(time.Date(2012, 12, 19, 0, 0, 0, 0, jst)))
	assert.Equal(t, e.URL.String(), "http://hatenablog.example.com/1")
	assert.Equal(t, e.EditURL, "http://hatenablog.example.com/1/edit")
	assert.Equal(t, e.Content, "test\ntest2\n")
}

var draftContent = `---
Title: 所内#4
Date: 2012-12-20T00:00:00+09:00
URL: http://hatenablog.example.com/2
EditURL: http://hatenablog.example.com/2/edit
Draft: true
---

下書き
`

func TestDraftFullContent(t *testing.T) {
	u, _ := url.Parse("http://hatenablog.example.com/2")
	jst, _ := time.LoadLocation("Asia/Tokyo")
	d := time.Date(2012, 12, 20, 0, 0, 0, 0, jst)

	e := &Entry{
		EntryHeader: &EntryHeader{
			URL:     &EntryURL{u},
			EditURL: u.String() + "/edit",
			Title:   "所内#4",
			Date:    &entryTime{&d},
			IsDraft: true,
		},
		LastModified: &d,
		Content:      "下書き\n",
	}
	assert.Equal(t, e.fullContent(), draftContent)
}

func TestFrontmatterDraftEntryFromReader(t *testing.T) {
	jst, _ := time.LoadLocation("Asia/Tokyo")

	e, err := entryFromReader(strings.NewReader(draftContent))
	assert.NoError(t, err)

	assert.Equal(t, e.Title, "所内#4")
	assert.True(t, e.Date.Equal(time.Date(2012, 12, 20, 0, 0, 0, 0, jst)))
	assert.Equal(t, e.URL.String(), "http://hatenablog.example.com/2")
	assert.Equal(t, e.EditURL, "http://hatenablog.example.com/2/edit")
	assert.True(t, e.IsDraft)
	assert.Equal(t, e.Content, "下書き\n")
}

var noCategory = `Title: 所内
Date: 2012-12-20T00:00:00+09:00
URL: http://hatenablog.example.com/2
EditURL: http://hatenablog.example.com/2/edit
`

func TestUnmarshalYAML(t *testing.T) {
	u, _ := url.Parse("http://hatenablog.example.com/2")
	jst, _ := time.LoadLocation("Asia/Tokyo")
	d := time.Date(2012, 12, 20, 0, 0, 0, 0, jst)

	eh := &EntryHeader{
		URL:      &EntryURL{u},
		EditURL:  u.String() + "/edit",
		Title:    "所内",
		Category: []string{"foo", "bar"},
		Date:     &entryTime{&d},
	}
	ya, _ := yaml.Marshal(eh)
	assert.Equal(t, `Title: 所内
Category:
- foo
- bar
Date: 2012-12-20T00:00:00+09:00
URL: http://hatenablog.example.com/2
EditURL: http://hatenablog.example.com/2/edit
`, string(ya))

	eh2 := EntryHeader{}
	yaml.Unmarshal(ya, &eh2)
	assert.Equal(t, "所内", eh2.Title)

	eh3 := EntryHeader{}
	yaml.Unmarshal([]byte(noCategory), &eh3)
	assert.Nil(t, eh3.Category)
}
