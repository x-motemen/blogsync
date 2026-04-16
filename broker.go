package main

import (
	"bufio"
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

// editURLFromFile extracts the EditURL from a file's frontmatter without
// fully parsing the entry. Returns "" if not found.
func editURLFromFile(fpath string) string {
	f, err := os.Open(fpath)
	if err != nil {
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	inFrontmatter := false
	for scanner.Scan() {
		line := scanner.Text()
		if line == "---" {
			if inFrontmatter {
				return ""
			}
			inFrontmatter = true
			continue
		}
		if !inFrontmatter {
			return ""
		}
		if strings.HasPrefix(line, "EditURL:") {
			return strings.TrimSpace(strings.TrimPrefix(line, "EditURL:"))
		}
	}
	return ""
}

// buildLocalEntryMap walks the local root and builds a map of EditURL to file path
// for all entry files.
func (b *broker) buildLocalEntryMap() map[string]string {
	m := map[string]string{}
	root := b.localRoot()
	filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, entryExt) {
			return nil
		}
		if editURL := editURLFromFile(path); editURL != "" {
			m[editURL] = path
		}
		return nil
	})
	return m
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
		if b.rootURL == "" {
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

	if e.IsDraft && e.isBlogEntry() {
		subdir, entryPath := b.blogConfig.extractEntryPath(e.URL.Path)
		if entryPath == "" {
			return ""
		}
		if isLikelyGivenPath(entryPath) {
			// EditURL is like bellow
			//   https://blog.hatena.ne.jp/Songmu/songmu.hatenadiary.org/atom/entry/6801883189050452361
			paths := strings.Split(e.EditURL, "/")
			if len(paths) == 8 {
				localPath = subdir + b.blogConfig.entryDirectory() + draftDir + paths[7] // path[7] is entryID
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

	if e.IsDraft && e.isBlogEntry() && e.URL != nil {
		// Clear temporary URL for entries stored in _draft/
		_, destEntryPath := b.blogConfig.extractEntryPath(path)
		if strings.HasPrefix(destEntryPath, draftDir) {
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

	// Force upload when CustomPath differs from current URL path
	if e.CustomPath != "" && re.URL != nil {
		_, currentEntryPath := b.blogConfig.extractEntryPath(re.URL.Path)
		if currentEntryPath != e.CustomPath {
			return true, b.PutEntry(e)
		}
		// CustomPath matches current URL — clear it to avoid unnecessary API update
		e.CustomPath = ""
	}

	if !newerWithAllowance(*e.LastModified, *re.LastModified) {
		return false, nil
	}

	return true, b.PutEntry(e)
}

func (b *broker) PutEntry(e *entry) error {
	newEntry, err := asEntry(b.Client.PutEntry(e.EditURL, e.atom()))
	if err != nil {
		return err
	}
	// Log URL change for published entries
	if !newEntry.IsDraft && e.URL != nil && newEntry.URL != nil && e.URL.Path != newEntry.URL.Path {
		logf("store", "URL changed: %s -> %s", e.URL.Path, newEntry.URL.Path)
	}
	// Preserve local path for drafts stored outside _draft/
	if e.localPath != "" && newEntry.IsDraft && newEntry.isBlogEntry() {
		_, entryPath := b.blogConfig.extractEntryPath(e.localPath)
		if entryPath != "" && !strings.HasPrefix(entryPath, draftDir) {
			newEntry.localPath = e.localPath
		}
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
	// Preserve local path for drafts stored outside _draft/
	if e.localPath != "" && newEntry.IsDraft && newEntry.isBlogEntry() {
		_, entryPath := b.blogConfig.extractEntryPath(e.localPath)
		if entryPath != "" && !strings.HasPrefix(entryPath, draftDir) {
			newEntry.localPath = e.localPath
		}
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
