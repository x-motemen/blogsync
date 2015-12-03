package main

import (
	"fmt"
	"io"
	"io/ioutil"
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
	Title      string     `yaml:"Title"`
	Date       *EntryTime `yaml:"Date"`
	URL        *EntryURL  `yaml:"URL"`
	EditURL    string     `yaml:"EditURL"`
	IsDraft    bool       `yaml:"Draft,omitempty"`
	CustomPath string     `yaml:"CustomPath,omitempty"`
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
	return strings.Join(headers, "\n") + "---\n\n"
}

func (e *Entry) fullContent() string {
	c := e.HeaderString() + e.Content
	if !strings.HasSuffix(c, "\n") {
		// fill newline for suppressing diff "No newline at end of file"
		c += "\n"
	}
	return c
}

func (e *Entry) atom() *atom.Entry {
	atomEntry := &atom.Entry{
		Title: e.Title,
		Content: atom.Content{
			Content: e.Content,
		},
	}
	if e.Date != nil {
		atomEntry.Updated = e.Date.Time
	}

	if e.IsDraft {
		atomEntry.Control = &atom.Control{
			Draft: "yes",
		}
	}
	if e.CustomPath != "" {
		atomEntry.CustomURL = e.CustomPath
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

var delimReg = regexp.MustCompile(`---\n+`)

func entryFromReader(source io.Reader, isNew bool) (*Entry, error) {
	b, err := ioutil.ReadAll(source)
	if err != nil {
		return nil, err
	}
	content := string(b)
	eh := EntryHeader{}
	if !isNew {
		c := delimReg.Split(content, 3)
		if len(c) != 3 || c[0] != "" {
			return nil, fmt.Errorf("entry format is invalid")
		}

		err = yaml.Unmarshal([]byte(c[1]), &eh)
		if err != nil {
			return nil, err
		}
		content = c[2]
	}
	entry := &Entry{
		EntryHeader: &eh,
		Content:     content,
	}

	if f, ok := source.(*os.File); ok {
		fi, err := os.Stat(f.Name())
		if err != nil {
			return nil, err
		}
		t := fi.ModTime()
		entry.LastModified = &t
	}

	return entry, nil
}

func asEntry(atomEntry *atom.Entry, err error) (*Entry, error) {
	if err != nil {
		return nil, err
	}

	return entryFromAtom(atomEntry)
}
