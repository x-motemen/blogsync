package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/motemen/blogsync/atom"
)

// Entry is an entry stored on remote blog providers
type Entry struct {
	URL          *url.URL
	Title        string
	Date         *time.Time
	EditURL      string
	LastModified *time.Time
	Content      string
	ContentType  string
}

func (e *Entry) HeaderString() string {
	return strings.Join([]string{
		"Title:   " + e.Title,
		"Date:    " + e.Date.Format(timeFormat),
		"URL:     " + e.URL.String(),
		"EditURL: " + e.EditURL,
	}, "\n") + "\n"
}

const timeFormat = "2006-01-02T15:04:05-07:00"

var rxHeader = regexp.MustCompile(`^(\w+):\s*(.+)`)

func (e *Entry) atom() *atom.Entry {
	return &atom.Entry{
		Title: e.Title,
		Content: atom.Content{
			Content: e.Content,
		},
		Updated: e.Date,
		XMLNs:   "http://www.w3.org/2005/Atom",
	}
}

func entryFromAtom(e *atom.Entry) (*Entry, error) {
	alternateLink := e.Links.Find("alternate")
	if alternateLink == nil {
		return nil, fmt.Errorf("could not find link[rel=alternate]")
	}

	u, err := url.Parse(alternateLink.Href)
	if err != nil {
		return nil, err
	}

	editLink := e.Links.Find("edit")
	if editLink == nil {
		return nil, fmt.Errorf("could not find link[rel=edit]")
	}

	return &Entry{
		URL:          u,
		EditURL:      editLink.Href,
		Title:        e.Title,
		Date:         e.Updated,
		LastModified: e.Edited,
		Content:      e.Content.Content,
		ContentType:  e.Content.Type,
	}, nil
}

func entryFromReader(source io.Reader) (*Entry, error) {
	r := bufio.NewReader(source)

	entry := &Entry{}

	if f, ok := source.(*os.File); ok {
		fi, err := os.Stat(f.Name())
		if err != nil {
			return nil, err
		}

		t := fi.ModTime()
		entry.LastModified = &t
	}

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
			t, err := time.Parse(timeFormat, value)
			if err != nil {
				return nil, err
			}
			entry.Date = &t
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

func asEntry(atomEntry *atom.Entry, err error) (*Entry, error) {
	if err != nil {
		return nil, err
	}

	return entryFromAtom(atomEntry)
}
