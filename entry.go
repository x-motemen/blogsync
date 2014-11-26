package main

import (
	"bufio"
	"bytes"
	"io"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

// RemoteEntry is an entry stored on remote blog providers
type RemoteEntry struct {
	URL          *url.URL
	Title        string
	Date         time.Time
	EditURL      string
	LastModified time.Time
	Content      string
	ContentType  string
}

func (re *RemoteEntry) HeaderString() string {
	return strings.Join([]string{
		"Title:   " + re.Title,
		"Date:    " + re.Date.Format(timeFormat),
		"URL:     " + re.URL.String(),
		"EditURL: " + re.EditURL,
	}, "\n") + "\n"
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

const timeFormat = "2006-01-02T15:04:05-07:00"

var rxHeader = regexp.MustCompile(`^(\w+):\s*(.+)`)

func EntryFromReader(r_ io.Reader) (*RemoteEntry, error) {
	r := bufio.NewReader(r_)

	entry := &RemoteEntry{}

	// TODO LastModified = mtime

	var body bytes.Buffer
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return nil, err
		}

		body.WriteString(line)

		m := rxHeader.FindStringSubmatch(line)
		if m == nil {
			if line == "\n" {
				// Discard lines so far because they are valid headers
				body.Reset()
			}
			break
		}

		key, value := m[1], m[2]
		switch key {
		case "Title":
			entry.Title = value
		case "Date":
			entry.Date, err = time.Parse(timeFormat, value)
			if err != nil {
				return nil, err
			}
		case "EditURL":
			entry.EditURL = value
		}
	}

	_, err := io.Copy(&body, r)
	if err != nil {
		return nil, err
	}

	entry.Content = body.String()

	return entry, nil
}
