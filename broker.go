package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/motemen/blogsync/atom"
	"github.com/motemen/go-wsse"
)

type broker struct {
	*atom.Client
	*blogConfig
}

func newBroker(bc *blogConfig) *broker {
	return &broker{
		Client: &atom.Client{
			Client: &http.Client{
				Transport: &wsse.Transport{
					Username: bc.Username,
					Password: bc.Password,
				},
			},
		},
		blogConfig: bc,
	}
}

func (b *broker) FetchRemoteEntries() ([]*Entry, error) {
	entries := []*Entry{}
	url := fmt.Sprintf("https://blog.hatena.ne.jp/%s/%s/atom/entry", b.Username, b.RemoteRoot)

	for {
		feed, err := b.Client.GetFeed(url)
		if err != nil {
			return nil, err
		}

		for _, ae := range feed.Entries {
			e, err := entryFromAtom(&ae)
			if err != nil {
				return nil, err
			}

			entries = append(entries, e)
		}

		nextLink := feed.Links.Find("next")
		if nextLink == nil {
			break
		}

		url = nextLink.Href
	}

	return entries, nil
}

func (b *broker) LocalPath(e *Entry) string {
	extension := ".md" // TODO regard re.ContentType
	return filepath.Join(b.LocalRoot, b.RemoteRoot, e.URL.Path+extension)
}

func (b *broker) StoreFresh(e *Entry, path string) (bool, error) {
	var localLastModified time.Time
	if fi, err := os.Stat(path); err == nil {
		localLastModified = fi.ModTime()
	}

	if e.LastModified.After(localLastModified) {
		logf("fresh", "remote=%s > local=%s", e.LastModified, localLastModified)
		if err := b.Store(e, path); err != nil {
			return false, err
		}

		return true, nil
	}

	return false, nil
}

func (b *broker) Store(e *Entry, path string) error {
	logf("store", "%s", path)

	dir, _ := filepath.Split(path)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}

	_, err = f.WriteString(e.fullContent())
	if err != nil {
		return err
	}

	err = f.Close()
	if err != nil {
		return err
	}

	return os.Chtimes(path, *e.LastModified, *e.LastModified)
}

func (b *broker) UploadFresh(e *Entry) (bool, error) {
	re, err := asEntry(b.Client.GetEntry(e.EditURL))
	if err != nil {
		return false, err
	}

	if e.LastModified.After(*re.LastModified) == false {
		return false, nil
	}

	return true, b.PutEntry(e)
}

func (b *broker) PutEntry(e *Entry) error {
	newEntry, err := asEntry(b.Client.PutEntry(e.EditURL, e.atom()))
	if err != nil {
		return err
	}
	if e.CustomPath != "" {
		newEntry.CustomPath = e.CustomPath
	}

	path := b.LocalPath(newEntry)
	return b.Store(newEntry, path)
}

func (b *broker) PostEntry(e *Entry) error {
	postURL := fmt.Sprintf("https://blog.hatena.ne.jp/%s/%s/atom/entry", b.Username, b.RemoteRoot)
	newEntry, err := asEntry(b.Client.PostEntry(postURL, e.atom()))
	if err != nil {
		return err
	}
	if e.CustomPath != "" {
		newEntry.CustomPath = e.CustomPath
	}

	path := b.LocalPath(newEntry)
	return b.Store(newEntry, path)
}
