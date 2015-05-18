package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/motemen/blogsync/atom"
	"gopkg.in/yaml.v2"
)

const timeFormat = "2006-01-02T15:04:05-07:00"

type EntryTime struct {
	*time.Time
}

type EntryURL struct {
	*url.URL
}

type EntryHeader struct {
	Title   string     `yaml:"Title"`
	Date    *EntryTime `yaml:"Date"`
	URL     *EntryURL  `yaml:"URL"`
	EditURL string     `yaml:"EditURL"`
	IsDraft bool       `yaml:"Draft,omitempty"`
}

func (eu *EntryURL) MarshalYAML() (interface{}, error) {
	return eu.String(), nil
}

func (et *EntryTime) MarshalYAML() (interface{}, error) {
	return et.Format(timeFormat), nil
}

func (eu *EntryURL) UnmarshalYAML(unmarshal func(v interface{}) error) error {
	var s string
	err := unmarshal(&s)
	if err != nil {
		return err
	}
	u, err := url.Parse(s)
	if err != nil {
		return err
	}
	eu.URL = u
	return nil
}

func (et *EntryTime) UnmarshalYAML(unmarshal func(v interface{}) error) error {
	var t time.Time
	err := unmarshal(&t)
	if err != nil {
		return err
	}
	et.Time = &t
	return nil
}

// Entry is an entry stored on remote blog providers
type Entry struct {
	*EntryHeader
	LastModified *time.Time
	Content      string
	ContentType  string
}

func (e *Entry) HeaderString() string {
	d, err := yaml.Marshal(e.EntryHeader)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	headers := []string{
		"---",
		string(d),
	}
	return strings.Join(headers, "\n") + "---\n"
}

func (e *Entry) fullContent() string {
	c := e.HeaderString() + e.Content
	if !strings.HasSuffix(c, "\n") {
		// fill newline for suppressing diff "No newline at end of file"
		c += "\n"
	}
	return c
}

var rxHeader = regexp.MustCompile(`^(?:\s*[*-]\s*)?(\w+):\s*(.+)`)

func (e *Entry) atom() *atom.Entry {
	atomEntry := &atom.Entry{
		Title: e.Title,
		Content: atom.Content{
			Content: e.Content,
		},
		Updated: e.Date.Time,
	}

	if e.IsDraft {
		atomEntry.Control = &atom.Control{
			Draft: "yes",
		}
	}

	return atomEntry
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

	entry := &Entry{
		EntryHeader: &EntryHeader{
			URL:     &EntryURL{u},
			EditURL: editLink.Href,
			Title:   e.Title,
			Date:    &EntryTime{e.Updated},
		},
		LastModified: e.Edited,
		Content:      e.Content.Content,
		ContentType:  e.Content.Type,
	}

	if e.Control != nil && e.Control.Draft == "yes" {
		entry.IsDraft = true
	}

	return entry, nil
}

func entryFromReader(source io.Reader) (*Entry, error) {
	r := bufio.NewReader(source)

	entry := &Entry{
		EntryHeader: &EntryHeader{},
	}

	if f, ok := source.(*os.File); ok {
		fi, err := os.Stat(f.Name())
		if err != nil {
			return nil, err
		}
		t := fi.ModTime()
		entry.LastModified = &t
	}

	var body bytes.Buffer
	lineNum := 0
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return nil, err
		}
		lineNum++
		if line == "---\n" && lineNum == 1 {
			continue
		}

		m := rxHeader.FindStringSubmatch(line)
		if m == nil {
			if line != "\n" && line != "---\n" {
				body.WriteString(line)
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
			entry.Date = &EntryTime{&t}
		case "EditURL":
			entry.EditURL = value
		case "Draft":
			entry.IsDraft = (value == "yes" || value == "true")
		case "URL":
			u, err := url.Parse(value)
			if err != nil {
				return nil, err
			}
			entry.URL = &EntryURL{u}
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
