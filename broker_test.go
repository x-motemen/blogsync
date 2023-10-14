package main

import (
	"net/url"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEntryEndPointUrl(t *testing.T) {
	testCases := []struct {
		name   string
		config blogConfig
		expect string
	}{
		{
			name: "username",
			config: blogConfig{
				BlogID:   "example1.hatenablog.com",
				Username: "sample1",
			},
			expect: "https://blog.hatena.ne.jp/sample1/example1.hatenablog.com/atom/entry",
		},
		{
			name: "owner",
			config: blogConfig{
				BlogID:   "example1.hatenablog.com",
				Username: "sample1",
				Owner:    "sample2",
			},
			expect: "https://blog.hatena.ne.jp/sample2/example1.hatenablog.com/atom/entry",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := entryEndPointUrl(&tc.config)
			assert.Equal(t, tc.expect, got)
		})
	}
}

func TestOriginalPath(t *testing.T) {
	u, _ := url.Parse("http://hatenablog.example.com/2")
	jst, _ := time.LoadLocation("Asia/Tokyo")
	d := time.Date(2023, 10, 10, 0, 0, 0, 0, jst)

	testCases := []struct {
		name          string
		entry         entry
		expect        string
		expectWindows string
	}{
		{
			name: "entry has URL",
			entry: entry{
				entryHeader: &entryHeader{
					URL:     &entryURL{u},
					EditURL: u.String() + "/edit",
					Title:   "test",
					Date:    &d,
					IsDraft: true,
				},
				LastModified: &d,
				Content:      "テスト",
			},
			expect:        "example1.hatenablog.com/2.md",
			expectWindows: "example1.hatenablog.com\\2.md",
		},
		{
			name: "Not URL",
			entry: entry{
				entryHeader: &entryHeader{
					EditURL: u.String() + "/edit",
					Title:   "hoge",
					IsDraft: true,
				},
				LastModified: &d,
				Content:      "テスト",
			},
			expect:        "",
			expectWindows: "",
		},
	}

	config := blogConfig{
		BlogID:   "example1.hatenablog.com",
		Username: "sample1",
	}
	broker := newBroker(&config)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := broker.originalPath(&tc.entry)
			if runtime.GOOS == "windows" {
				tc.expect = tc.expectWindows
			}
			assert.Equal(t, tc.expect, got)
		})
	}
}
