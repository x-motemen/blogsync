package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/motemen/blogsync/atom"
)

type Broker struct {
	client *WSSEClient
	*BlogConfig
}

func NewBroker(c *BlogConfig) *Broker {
	return &Broker{
		client: &WSSEClient{
			Client:   http.DefaultClient,
			UserName: c.UserName,
			Password: c.Password,
		},
		BlogConfig: c,
	}
}

func (b *Broker) FetchRemoteEntries() ([]*Entry, error) {
	resp, err := b.client.Get(fmt.Sprintf("https://blog.hatena.ne.jp/%s/%s/atom/entry", b.UserName, b.RemoteRoot))
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

	return b.RemoteEntriesFromFeed(feed)
}

func entryFromAtom(e *atom.Entry) (*Entry, error) {
	u, err := url.Parse(atom.FindLink("alternate", e.Links).Href)
	if err != nil {
		return nil, err
	}

	editLink := atom.FindLink("edit", e.Links)
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

func (b *Broker) RemoteEntriesFromFeed(feed *atom.Feed) ([]*Entry, error) {
	remoteEntries := make([]*Entry, len(feed.Entries))

	for i, e := range feed.Entries {
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
	var entryXML bytes.Buffer

	atomEntry := atom.Entry{
		Title: e.Title,
		Content: atom.Content{
			Content: e.Content,
		},
		Updated: e.Date,
		XMLNs:   "http://www.w3.org/2005/Atom",
	}

	entryXML.WriteString(xml.Header)
	enc := xml.NewEncoder(&entryXML)
	enc.Indent("", "  ")
	enc.Encode(atomEntry)

	resp, err := b.client.Put(e.EditURL, &entryXML)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		bytes, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("got [%s]: %q", resp.Status, string(bytes))
	}

	resultAtomEntry, err := atom.ParseEntry(resp.Body)
	if err != nil {
		return err
	}

	resultEntry, err := entryFromAtom(resultAtomEntry)
	if err != nil {
		return err
	}

	path := b.LocalPath(resultEntry)
	return b.Download(resultEntry, path)
}
