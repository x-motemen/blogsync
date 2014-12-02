package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/motemen/blogsync/atom"
	"github.com/motemen/go-wsse"
)

type Broker struct {
	client *atom.Client
	*BlogConfig
}

func NewBroker(c *BlogConfig) *Broker {
	return &Broker{
		client: &atom.Client{
			Client: &http.Client{
				Transport: &wsse.Transport{
					Username: c.Username,
					Password: c.Password,
				},
			},
		},
		BlogConfig: c,
	}
}

func (b *Broker) FetchRemoteEntries() ([]*Entry, error) {
	entries := []atom.Entry{}
	url := fmt.Sprintf("https://blog.hatena.ne.jp/%s/%s/atom/entry", b.Username, b.RemoteRoot)

	for {
		resp, err := b.client.Get(url)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != 200 {
			return nil, fmt.Errorf("request not succeeded: got [%s]", resp.Status)
		}

		feed, err := atom.Parse(resp.Body)
		if err != nil {
			return nil, err
		}

		entries = append(entries, feed.Entries...)

		nextLink := feed.Links.Find("next")
		if nextLink == nil {
			break
		}

		url = nextLink.Href
	}

	return b.entriesFromAtomEntries(entries)
}

func entryFromAtom(e *atom.Entry) (*Entry, error) {
	u, err := url.Parse(e.Links.Find("alternate").Href)
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

func (b *Broker) entriesFromAtomEntries(entries []atom.Entry) ([]*Entry, error) {
	remoteEntries := make([]*Entry, len(entries))

	for i, e := range entries {
		re, err := entryFromAtom(&e)
		if err != nil {
			return nil, err
		}

		remoteEntries[i] = re
	}

	return remoteEntries, nil
}

func (b *Broker) LocalPath(e *Entry) string {
	extension := ".md" // TODO regard re.ContentType
	return filepath.Join(b.LocalRoot, e.URL.Host, e.URL.Path+extension)
}

func (b *Broker) Mirror(re *Entry, path string) (bool, error) {
	var localLastModified time.Time
	if fi, err := os.Stat(path); err == nil {
		localLastModified = fi.ModTime()
	}

	if re.LastModified.After(localLastModified) {
		logf("fresh", "remote=%s > local=%s", re.LastModified, localLastModified)
		if err := b.Download(re, path); err != nil {
			return false, err
		}

		return true, nil
	}

	return false, nil
}

func (b *Broker) Download(re *Entry, path string) error {
	logf("download", "%s -> %s", re.URL, path)

	dir, _ := filepath.Split(path)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}

	_, err = f.WriteString(re.HeaderString() + "\n" + re.Content)
	if err != nil {
		return err
	}

	err = f.Close()
	if err != nil {
		return err
	}

	return os.Chtimes(path, re.LastModified, re.LastModified)
}

func (b *Broker) Upload(e *Entry) (bool, error) {
	resp, err := b.client.Get(e.EditURL)
	if err != nil {
		return false, err
	}

	atomEntry, err := atom.ParseEntry(resp.Body)
	if err != nil {
		return false, err
	}

	// TODO Entry でほしい
	if e.LastModified.After(atomEntry.Edited) == false {
		return false, nil
	}

	return true, b.Put(e)

	// TODO put 後にローカル書き換える
}

func (b *Broker) Put(e *Entry) error {
	atomEntry := atom.Entry{
		Title: e.Title,
		Content: atom.Content{
			Content: e.Content,
		},
		Updated: e.Date,
		XMLNs:   "http://www.w3.org/2005/Atom",
	}

	newAtomEntry, err := b.client.PutEntry(e.EditURL, &atomEntry)
	if err != nil {
		return err
	}

	newEntry, err := entryFromAtom(newAtomEntry)
	if err != nil {
		return err
	}

	path := b.LocalPath(newEntry)
	return b.Download(newEntry, path)
}
