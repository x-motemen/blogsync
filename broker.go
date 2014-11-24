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
	"strings"

	"github.com/motemen/blogsync/atom"
)

type Broker struct {
	*BlogConfig
}

func NewBroker(c *BlogConfig) *Broker {
	return &Broker{
		BlogConfig: c,
	}
}

func (b *Broker) client() *WSSEClient {
	return &WSSEClient{
		Client:   http.DefaultClient,
		UserName: b.UserName,
		Password: b.Password,
	}
}

func (b *Broker) FetchRemoteEntries() ([]*RemoteEntry, error) {
	resp, err := b.client().Get(fmt.Sprintf("https://blog.hatena.ne.jp/%s/%s/atom/entry", b.UserName, b.RemoteRoot))
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

func (b *Broker) RemoteEntriesFromFeed(feed *atom.Feed) ([]*RemoteEntry, error) {
	remoteEntries := make([]*RemoteEntry, len(feed.Entries))

	for i, e := range feed.Entries {
		u, err := url.Parse(atom.FindLink("alternate", e.Links).Href)
		if err != nil {
			return nil, err
		}

		remoteEntries[i] = &RemoteEntry{
			URL:         u,
			EntryID:     e.ID,
			Title:       e.Title,
			Date:        e.Updated,
			Content:     e.Content.Content,
			ContentType: e.Content.Type,
		}
	}

	return remoteEntries, nil
}

func (b *Broker) LocalHalf(re *RemoteEntry) *LocalEntry {
	extension := ".md" // TODO regard re.ContentType
	path := filepath.Join(b.LocalRoot, re.URL.Host, re.URL.Path+extension)
	return &LocalEntry{
		Path: path,
	}
}

func (b *Broker) Download(re *RemoteEntry, le *LocalEntry) error {
	logf("download", "%s -> %s", re.URL, le.Path)

	dir, _ := filepath.Split(le.Path)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}

	f, err := os.Create(le.Path)
	if err != nil {
		return err
	}

	_, err = f.WriteString(re.HeaderString() + "\n" + re.Content)
	if err != nil {
		return err
	}

	return nil
}

func (b *Broker) Put(e *RemoteEntry) error {
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

	// XXX workaround
	parts := strings.Split(e.EntryID, "-")
	entryID := parts[len(parts)-1]

	putReq, err := http.NewRequest(
		"PUT",
		fmt.Sprintf(
			"https://blog.hatena.ne.jp/%s/%s/atom/entry/%s",
			b.UserName, b.RemoteRoot, entryID,
		),
		&entryXML,
	)
	if err != nil {
		return err
	}

	logf("xml", "%s", entryXML.String())

	logf("upload", "x -> x")
	resp, err := b.client().Do(putReq)

	bytes, _ := ioutil.ReadAll(resp.Body)
	logf("error", "%s", string(bytes))

	return err
}
