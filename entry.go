package main

import (
	"net/url"
	"os"
	"strings"
	"time"
)

// RemoteEntry is an entry stored on remote blog providers
type RemoteEntry struct {
	URL         *url.URL
	Title       string
	Date        time.Time
	Path        string
	Content     string
	ContentType string
}

func (re *RemoteEntry) HeaderString() string {
	return strings.Join([]string{
		"Title: " + re.Title,
		"Date:  " + re.Date.Format("2006-01-02T15:04:05"),
	}, "\n") + "\n"
}

func (re *RemoteEntry) LastModified() time.Time {
	// XXX はてなブログの Atom において lastModified 的なものはない気がする
	return time.Now()
}

type LocalEntry struct {
	Path string
}

var oldest = time.Unix(0, 0)

func (le *LocalEntry) LastModified() time.Time {
	fi, err := os.Stat(le.Path)
	if err != nil {
		return oldest
	}
	return fi.ModTime()
}
