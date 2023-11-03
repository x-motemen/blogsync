package main

import (
	"testing"
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
			if tc.expect != got {
				t.Errorf("expect: %s, got: %s", tc.expect, got)
			}
		})
	}
}
