package main

import (
	"net/url"
	"os"
	"strings"
	"testing"
	"time"
)

var jst *time.Location

func init() {
	jst, _ = time.LoadLocation("Asia/Tokyo")
}

func TestEntryFromReader(t *testing.T) {
	f, err := os.Open("example/data/karimen.hatenablog.com/entry/2012/12/18/000000.md")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	e, err := entryFromReader(f)
	if err != nil {
		t.Errorf("failed to parse entry: %v", err)
	}
	if e, g := "所内#2", e.Title; e != g {
		t.Errorf("Title: got %#v, want %#v", g, e)
	}

	if e, g := e.Date, time.Date(2012, 12, 18, 0, 0, 0, 0, jst); !e.Equal(g) {
		t.Errorf("Date: got %#v, want %#v", g, e)
	}
}

func mustURLParse(s string) *url.URL {
	u, err := url.Parse(s)
	if err != nil {
		panic(err)
	}
	return u
}

func TestFullContent(t *testing.T) {
	testCases := []struct {
		name string
		e    func() *entry
		want string
	}{{
		name: "normal",
		e: func() *entry {
			u := mustURLParse("http://hatenablog.example.com/1")
			d := time.Date(2012, 12, 19, 0, 0, 0, 0, jst)
			return &entry{
				entryHeader: &entryHeader{
					URL:     &entryURL{u},
					EditURL: u.String() + "/edit",
					Title:   "所内#3",
					Date:    &d,
				},
				LastModified: &d,
				Content:      "test\ntest2\n",
			}
		},
		want: `---
Title: 所内#3
Date: 2012-12-19T00:00:00+09:00
URL: http://hatenablog.example.com/1
EditURL: http://hatenablog.example.com/1/edit
---

test
test2
`}, {
		name: "draft",
		e: func() *entry {
			u := mustURLParse("http://hatenablog.example.com/2")
			return &entry{
				entryHeader: &entryHeader{
					URL:     &entryURL{u},
					EditURL: u.String() + "/edit",
					Title:   "所内#4",
					IsDraft: true,
				},
				Content: "下書き\n",
			}

		},
		want: `---
Title: 所内#4
URL: http://hatenablog.example.com/2
EditURL: http://hatenablog.example.com/2/edit
Draft: true
---

下書き
`}, {
		name: "draft with future date",
		e: func() *entry {
			u := mustURLParse("http://hatenablog.example.com/2")
			d := time.Date(2099, 12, 20, 0, 0, 0, 0, jst)

			return &entry{
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

		},
		want: `---
Title: 所内#4
Date: 2099-12-20T00:00:00+09:00
URL: http://hatenablog.example.com/2
EditURL: http://hatenablog.example.com/2/edit
Draft: true
---

下書き
`}, {
		name: "category",
		e: func() *entry {
			u := mustURLParse("http://hatenablog.example.com/2")
			d := time.Date(2012, 12, 20, 0, 0, 0, 0, jst)

			return &entry{
				entryHeader: &entryHeader{
					URL:      &entryURL{u},
					EditURL:  u.String() + "/edit",
					Title:    "所内",
					Category: []string{"foo", "bar"},
					Date:     &d,
				},
				LastModified: &d,
				Content:      "foo bar カテゴリー\n",
			}

		},
		want: `---
Title: 所内
Category:
- foo
- bar
Date: 2012-12-20T00:00:00+09:00
URL: http://hatenablog.example.com/2
EditURL: http://hatenablog.example.com/2/edit
---

foo bar カテゴリー
`}}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := tc.e()
			fullContent := e.fullContent()
			if fullContent != tc.want {
				t.Errorf("got %v, want %v", fullContent, tc.want)
			}

			parsedE, err := entryFromReader(strings.NewReader(fullContent))
			if err != nil {
				t.Errorf("failed to parse entry: %v", err)
			}

			if e, g := e.Title, parsedE.Title; e != g {
				t.Errorf("Title: got %#v, want %#v", g, e)
			}

			if e.Date != nil {
				if e, g := e.Date, parsedE.Date; !e.Equal(*g) {
					t.Errorf("Date: got %#v, want %#v", g, e)
				}
			}
			if e, g := e.URL, parsedE.URL; e.String() != g.String() {
				t.Errorf("URL: got %#v, want %#v", g, e)
			}
			if e, g := e.EditURL, parsedE.EditURL; e != g {
				t.Errorf("EditURL: got %#v, want %#v", g, e)
			}
			if e, g := e.Content, parsedE.Content; e != g {
				t.Errorf("Content: got %#v, want %#v", g, e)
			}
			if e, g := e.IsDraft, parsedE.IsDraft; e != g {
				t.Errorf("IsDraft: got %#v, want %#v", g, e)
			}
		})
	}
}
