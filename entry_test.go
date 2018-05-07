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

	assert.Equal(t, "所内#2", e.Title)

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

	e := &entry{
		entryHeader: &entryHeader{
			URL:     &entryURL{u},
			EditURL: u.String() + "/edit",
			Title:   "所内#3",
			Date:    &d,
		},
		LastModified: &d,
		Content:      "test\ntest2",
	}
	assert.Equal(t, content, e.fullContent())
}

func TestFrontmatterEntryFromReader(t *testing.T) {
	jst, _ := time.LoadLocation("Asia/Tokyo")

	e, err := entryFromReader(strings.NewReader(content))
	assert.NoError(t, err)

	assert.Equal(t, "所内#3", e.Title)
	assert.True(t, e.Date.Equal(time.Date(2012, 12, 19, 0, 0, 0, 0, jst)))
	assert.Equal(t, "http://hatenablog.example.com/1", e.URL.String())
	assert.Equal(t, "http://hatenablog.example.com/1/edit", e.EditURL)
	assert.Equal(t, "test\ntest2\n", e.Content)
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

	e := &entry{
		entryHeader: &entryHeader{
			URL:     &entryURL{u},
			EditURL: u.String() + "/edit",
			Title:   "所内#4",
			Date:    &d,
			IsDraft: true,
		},
		LastModified: &d,
		Content:      "下書き\n",
	}
	assert.Equal(t, draftContent, e.fullContent())
}

func TestFrontmatterDraftEntryFromReader(t *testing.T) {
	jst, _ := time.LoadLocation("Asia/Tokyo")

	e, err := entryFromReader(strings.NewReader(draftContent))
	assert.NoError(t, err)

	assert.Equal(t, "所内#4", e.Title)
	assert.True(t, e.Date.Equal(time.Date(2012, 12, 20, 0, 0, 0, 0, jst)))
	assert.Equal(t, "http://hatenablog.example.com/2", e.URL.String())
	assert.Equal(t, "http://hatenablog.example.com/2/edit", e.EditURL)
	assert.True(t, e.IsDraft)
	assert.Equal(t, "下書き\n", e.Content)
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

	eh := &entryHeader{
		URL:      &entryURL{u},
		EditURL:  u.String() + "/edit",
		Title:    "所内",
		Category: []string{"foo", "bar"},
		Date:     &d,
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

	eh2 := entryHeader{}
	yaml.Unmarshal(ya, &eh2)
	assert.Equal(t, "所内", eh2.Title)

	eh3 := entryHeader{}
	yaml.Unmarshal([]byte(noCategory), &eh3)
	assert.Nil(t, eh3.Category)
}
