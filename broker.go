package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/motemen/go-wsse"
	"github.com/x-motemen/blogsync/atom"
)

type broker struct {
	*atom.Client
	*blogConfig
	writer io.Writer
}

func newBroker(bc *blogConfig, w io.Writer) *broker {
	if w == nil {
		w = os.Stdout
	}
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
		writer:     w,
	}
}

func (b *broker) FetchRemoteEntries(published, drafts bool) ([]*entry, error) {
	entries := []*entry{}
	staticPageURL := staticPageEndpointURL(b.blogConfig)
	urls := []string{
		entryEndPointUrl(b.blogConfig),
		staticPageURL,
	}
	for url := ""; true; {
		if url == "" {
			if len(urls) == 0 {
				break
			}
			url, urls = urls[0], urls[1:]
		}

		feed, err := b.Client.GetFeed(url)
		if err != nil {
			if url == staticPageURL {
				// Ignore errors in the case of static pages, because static page is the feature
				// only for pro users.
				break
			}
			return nil, err
		}
		if b.rootURL != "" {
			if l := feed.Links.Find("alternate"); l != nil {
				b.rootURL = l.Href
			}
		}

		for _, ae := range feed.Entries {
			e, err := entryFromAtom(&ae)
			if err != nil {
				return nil, err
			}
			if (e.IsDraft && drafts) || (!e.IsDraft && published) {
				entries = append(entries, e)
			}
		}

		nextLink := feed.Links.Find("next")
		if nextLink != nil {
			url = nextLink.Href
		} else {
			url = ""
		}
	}

	return entries, nil
}

const entryExt = ".md" // TODO regard re.ContentType

func (b *broker) LocalPath(e *entry) string {
	if e.localPath != "" {
		return e.localPath
	}
	if e.URL == nil {
		return ""
	}
	localPath := e.URL.Path

	if e.IsDraft && strings.Contains(e.EditURL, "/atom/entry/") {
		subdir, entryPath := extractEntryPath(e.URL.Path)
		if entryPath == "" {
			return ""
		}
		if isLikelyGivenPath(entryPath) {
			// EditURL is like bellow
			//   https://blog.hatena.ne.jp/Songmu/songmu.hatenadiary.org/atom/entry/6801883189050452361
			paths := strings.Split(e.EditURL, "/")
			if len(paths) == 8 {
				localPath = subdir + "/entry/" + draftDir + paths[7] // path[7] is entryID
			}
		}
	}
	return filepath.Join(b.localRoot(), localPath+entryExt)
}

func (b *broker) StoreFresh(e *entry, path string) (bool, error) {
	localLastModified, _ := modTime(path)
	if e.LastModified.After(localLastModified) {
		logf("fresh", "remote=%s > local=%s", e.LastModified, localLastModified)
		if err := b.Store(e, path, ""); err != nil {
			return false, err
		}

		return true, nil
	}

	return false, nil
}

func (b *broker) Store(e *entry, path, origPath string) error {
	logf("store", "%s", path)

	if e.IsDraft && strings.Contains(e.EditURL, "/atom/entry/") {
		_, entryPath := extractEntryPath(e.URL.Path)
		if entryPath == "" {
			return fmt.Errorf("invalid path: %s", e.URL.Path)
		}
		if isLikelyGivenPath(entryPath) {
			e.URL = nil
		}
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	if err := os.WriteFile(path, []byte(e.fullContent()), 0666); err != nil {
		return err
	}
	fmt.Fprintln(b.writer, path)

	if err := os.Chtimes(path, *e.LastModified, *e.LastModified); err != nil {
		return err
	}

	if origPath != "" && path != origPath {
		return os.Remove(origPath)
	}

	return nil
}

func (b *broker) UploadFresh(e *entry) (bool, error) {
	re, err := asEntry(b.Client.GetEntry(e.EditURL))
	if err != nil {
		return false, err
	}

	if e.LastModified.After(*re.LastModified) == false {
		return false, nil
	}

	return true, b.PutEntry(e)
}

func (b *broker) PutEntry(e *entry) error {
	newEntry, err := asEntry(b.Client.PutEntry(e.EditURL, e.atom()))
	if err != nil {
		return err
	}
	return b.Store(newEntry, b.LocalPath(newEntry), b.LocalPath(e))
}

func (b *broker) PostEntry(e *entry, isPage bool) error {
	var endPoint string
	if !isPage {
		endPoint = entryEndPointUrl(b.blogConfig)
	} else {
		endPoint = staticPageEndpointURL(b.blogConfig)
	}
	newEntry, err := asEntry(b.Client.PostEntry(endPoint, e.atom()))
	if err != nil {
		return err
	}
	return b.Store(newEntry, b.LocalPath(newEntry), "")
}

func (b *broker) RemoveEntry(e *entry) error {
	err := b.Client.DeleteEntry(e.EditURL)
	if err != nil {
		return err
	}
	p := b.LocalPath(e)
	if _, err := os.Stat(p); err == nil {
		return os.Remove(p)
	}
	return nil
}

func atomEndpointURLRoot(bc *blogConfig) string {
	owner := bc.Owner
	if owner == "" {
		owner = bc.Username
	}
	return fmt.Sprintf("https://blog.hatena.ne.jp/%s/%s/atom/", owner, bc.BlogID)
}

func entryEndPointUrl(bc *blogConfig) string {
	return atomEndpointURLRoot(bc) + "entry"
}

func staticPageEndpointURL(bc *blogConfig) string {
	return atomEndpointURLRoot(bc) + "page"
}
