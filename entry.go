package main

import (
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/x-motemen/blogsync/atom"
	"gopkg.in/yaml.v2"
)

const timeFormat = "2006-01-02T15:04:05-07:00"

type entryURL struct {
	*url.URL
}

type entryHeader struct {
	Title      string     `yaml:"Title"`
	Category   []string   `yaml:"Category,omitempty"`
	Date       *time.Time `yaml:"Date,omitempty"`
	URL        *entryURL  `yaml:"URL"`
	EditURL    string     `yaml:"EditURL"`
	PreviewURL string     `yaml:"PreviewURL,omitempty"`
	IsDraft    bool       `yaml:"Draft,omitempty"`
	CustomPath string     `yaml:"CustomPath,omitempty"`
}

func (eu *entryURL) MarshalYAML() (interface{}, error) {
	return eu.String(), nil
}

func (eu *entryURL) UnmarshalYAML(unmarshal func(v interface{}) error) error {
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

func (eh *entryHeader) blogID() (string, error) {
	// EditURL: https://blog.hatena.ne.jp/Songmu/songmu.hateblog.jp/atom/entry/...
	// "songmu.hateblog.jp" is blogID in above case.
	paths := strings.Split(eh.EditURL, "/")
	if len(paths) < 5 {
		return "", fmt.Errorf("failed to get blogID form EditURL: %s", eh.EditURL)
	}
	return paths[4], nil
}

// Entry is an entry stored on remote blog providers
type entry struct {
	*entryHeader
	LastModified *time.Time
	Content      string
	ContentType  string
}

func (e *entry) HeaderString() string {
	d, err := yaml.Marshal(e.entryHeader)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	headers := []string{
		"---",
		string(d),
	}
	return strings.Join(headers, "\n") + "---\n\n"
}

func (e *entry) fullContent() string {
	c := e.HeaderString() + e.Content
	if !strings.HasSuffix(c, "\n") {
		// fill newline for suppressing diff "No newline at end of file"
		c += "\n"
	}
	return c
}

func (e *entry) atom() *atom.Entry {
	atomEntry := &atom.Entry{
		Title: e.Title,
		Content: atom.Content{
			Content: e.Content,
		},
	}

	categories := make([]atom.Category, 0)
	for _, c := range e.Category {
		categories = append(categories, atom.Category{Term: c})
	}
	atomEntry.Category = categories

	if e.Date != nil {
		atomEntry.Updated = e.Date
	}

	if e.IsDraft {
		atomEntry.Control = &atom.Control{
			Draft:   "yes",
			Preview: "yes",
		}
	}

	if e.CustomPath != "" {
		atomEntry.CustomURL = e.CustomPath
	}

	return atomEntry
}

func entryFromAtom(e *atom.Entry) (*entry, error) {
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

	var previewLink string
	p := e.Links.Find("preview")
	if p != nil {
		previewLink = p.Href
	}

	categories := make([]string, 0)
	for _, c := range e.Category {
		categories = append(categories, c.Term)
	}

	var isDraft bool
	if e.Control != nil && e.Control.Draft == "yes" {
		isDraft = true
	}

	// Set the updated to nil when the entry is still draft.
	// But if the date is in the future, don't set to nil because it may be a reserved post.
	updated := e.Updated
	if updated != nil && isDraft && !time.Now().Before(*updated) {
		updated = nil
	}

	return &entry{
		entryHeader: &entryHeader{
			URL:        &entryURL{u},
			EditURL:    editLink.Href,
			PreviewURL: previewLink,
			Title:      e.Title,
			Category:   categories,
			Date:       updated,
			IsDraft:    isDraft,
		},
		LastModified: e.Edited,
		Content:      e.Content.Content,
		ContentType:  e.Content.Type,
	}, nil
}

var delimReg = regexp.MustCompile(`---\n+`)

func entryFromReader(source io.Reader) (*entry, error) {
	b, err := io.ReadAll(source)
	if err != nil {
		return nil, err
	}
	content := string(b)
	isNew := !strings.HasPrefix(content, "---\n")
	eh := entryHeader{}
	if !isNew {
		c := delimReg.Split(content, 3)
		if len(c) != 3 || c[0] != "" {
			return nil, fmt.Errorf("entry format is invalid")
		}

		err = yaml.Unmarshal([]byte(c[1]), &eh)
		if err != nil {
			return nil, err
		}
		// Set the updated to nil when the entry is still draft.
		// But if the date is in the future, don't set to nil because it may be a reserved post.
		if eh.IsDraft && eh.Date != nil && !time.Now().Before(*eh.Date) {
			eh.Date = nil
		}
		content = c[2]
	}
	entry := &entry{
		entryHeader: &eh,
		Content:     content,
	}

	if f, ok := source.(*os.File); ok {
		if runtime.GOOS == "windows" && f.Name() == os.Stdin.Name() {
			t := time.Now()
			entry.LastModified = &t
		} else {
			fi, err := os.Stat(f.Name())
			if err != nil {
				return nil, err
			}
			t := fi.ModTime()
			entry.LastModified = &t
		}
	}

	return entry, nil
}

func asEntry(atomEntry *atom.Entry, err error) (*entry, error) {
	if err != nil {
		return nil, err
	}

	return entryFromAtom(atomEntry)
}
