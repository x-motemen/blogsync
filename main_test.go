//go:build darwin || integration

package main

import (
	"bytes"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/urfave/cli/v2"
)

func blogsyncApp(app *cli.App) func(...string) (string, error) {
	buf := &bytes.Buffer{}
	app.Writer = buf
	return func(args ...string) (string, error) {
		buf.Reset()
		err := app.Run(append([]string{""}, args...))
		return strings.TrimSpace(buf.String()), err
	}
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

var draftFileReg = regexp.MustCompile(`entry/_draft/\d+\.md$`)

func TestBlogsync(t *testing.T) {
	blogID := os.Getenv("BLOGSYNC_TEST_BLOG")
	if blogID == "" {
		t.Skip("BLOGSYNC_TEST_BLOG not set")
	}

	pwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(pwd); err != nil {
			t.Fatal(err)
		}
	}()

	dir, err = filepath.EvalSymlinks(dir)
	if err != nil {
		t.Fatal(err)
	}

	confYAML := fmt.Sprintf(`%s:
  local_root: .
  omit_domain: true
`, blogID)
	if owner := os.Getenv("BLOGSYNC_OWNER"); owner != "" {
		confYAML += fmt.Sprintf("  owner: %s\n", owner)
	}
	conf := filepath.Join(dir, "blogsync.yaml")
	if err := os.WriteFile(conf, []byte(confYAML), 0644); err != nil {
		t.Fatal(err)
	}

	app := newApp()
	blogsync := blogsyncApp(app)

	t.Run("pull", func(t *testing.T) {
		if _, err := blogsync("pull"); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("list", func(t *testing.T) {
		if _, err := blogsync("list"); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("fetchRootURL", func(t *testing.T) {
		t.Log("fetchRootURL returns the remote root URL")
		conf, err := loadConfiguration()
		if err != nil {
			t.Fatal(err)
		}
		remotRoot := conf.Get(blogID).fetchRootURL()
		u, _ := url.Parse(remotRoot)
		// XXX: In case of own URL, blogID and remote URL do not match.
		if u.Hostname() != blogID {
			t.Errorf("unexpected hostname. got: %s, expected: %s", u.Hostname(), blogID)
		}
	})

	t.Run("post draft and publish", func(t *testing.T) {
		t.Log("Post a draft without a custom path and check if the file is saved in the proper location")
		app.Reader = strings.NewReader("draft\n")
		entryFile, err := blogsync("post", "--draft", blogID)
		app.Reader = os.Stdin
		if err != nil {
			t.Fatal(err)
		}
		defer func() {
			t.Log("remove the published entry")
			if _, err := blogsync("remove", entryFile); err != nil {
				t.Fatal(err)
			}
		}()

		if !draftFileReg.MatchString(entryFile) {
			t.Fatalf("unexpected draft file: %s", entryFile)
		}

		t.Log("Check entry metadata")
		e, err := entryFromFile(entryFile)
		if err != nil {
			t.Fatal(err)
		}
		if e.URL != nil {
			t.Errorf("URL is registered in a draft with no custom path specified. URL: %s", *e.URL)
		}
		if e.Date != nil {
			t.Errorf("Date is registered in a draft. Date: %s", *e.Date)
		}

		t.Log("Draft files under `_draft/` will revert to the original file name if the file is renamed and pushed again")
		d, f := filepath.Split(entryFile)
		movedPath := filepath.Join(d, "_"+f)
		if err := os.Rename(entryFile, movedPath); err != nil {
			t.Fatal(err)
		}
		originalEntryFile := entryFile
		entryFile = movedPath
		if err := appendFile(movedPath, "updated\n"); err != nil {
			t.Fatal(err)
		}
		draftFile, err := blogsync("push", entryFile)
		if err != nil {
			t.Fatal(err)
		}
		if draftFile != originalEntryFile {
			entryFile = draftFile
			t.Fatalf("unexpected draft file: %s", draftFile)
		}
		if exists(entryFile) {
			t.Errorf("renamed draft file is not deleted: %s", movedPath)
		}
		entryFile = draftFile

		t.Log("Check if the draft is fetched and saved in the proper location")
		if _, err := blogsync("fetch", entryFile); err != nil {
			t.Error(err)
		}

		t.Log("When a draft is published, a URL is issued and the file is saved in the corresponding location")
		publishedFile, err := blogsync("push", "--publish", entryFile)
		if err != nil {
			t.Fatal(err)
		}
		if exists(entryFile) {
			t.Errorf("draft file not deleted: %s", entryFile)
		}
		entryFile = publishedFile

		_, entryPath := extractEntryPath(entryFile)
		if !isLikelyGivenPath(entryPath) {
			t.Errorf("unexpected published file: %s", entryFile)
		}
	})

	t.Run("post draft and publish with custom path", func(t *testing.T) {
		t.Log("Creating a draft with a custom path saves the file in the specified location, not under `_draft/`")
		localFile := filepath.Join(dir, "entry", time.Now().Format("custom-20060102150405")+".md")
		if err := os.WriteFile(localFile, []byte(`---
Draft: true
---
test`), 0644); err != nil {
			t.Fatal(err)
		}
		entryFile, err := blogsync("push", localFile)
		if err != nil {
			t.Fatal(err)
		}
		defer func() {
			t.Log("remove the published entry")
			if _, err := blogsync("remove", entryFile); err != nil {
				t.Fatal(err)
			}
		}()
		if entryFile != localFile {
			t.Errorf("unexpected published file: %s", entryFile)
		}

		t.Log("When publishing a draft with a custom path, the file location is unchanged")
		publishedFile, err := blogsync("push", "--publish", entryFile)
		if err != nil {
			t.Fatal(err)
		}
		if publishedFile != entryFile {
			t.Errorf("unexpected published file: %s", publishedFile)
		}

		t.Log("If the file name of an entry is changed, the custom path will follow suit")
		d, f := filepath.Split(entryFile)
		movedPath := filepath.Join(d, "custom-"+f)
		if err := os.Rename(entryFile, movedPath); err != nil {
			t.Fatal(err)
		}
		entryFile = movedPath
		if err := appendFile(entryFile, "updated\n"); err != nil {
			t.Fatal(err)
		}
		publishedFile, err = blogsync("push", entryFile)
		if err != nil {
			t.Fatal(err)
		}
		if publishedFile != entryFile {
			entryFile = publishedFile
			t.Errorf("unexpected published file: %s", publishedFile)
		}
	})

	t.Run("post static pages", func(t *testing.T) {
		t.Log("Posting a static page without a custom path results in an error")
		app.Reader = strings.NewReader("static\n")
		if _, err := blogsync("post", "--page", blogID); err == nil {
			t.Error("expected error did not occur")
		}

		t.Log("Posting a static page with a custom path saves the file in the specified location")
		customPath := fmt.Sprintf("static-%s", time.Now().Format("20060102150405"))
		app.Reader = strings.NewReader("static\n")
		entryFile, err := blogsync("post", "--page", "--custom-path", customPath, blogID)
		app.Reader = os.Stdin
		if err != nil {
			t.Fatal(err)
		}
		if entryFile != filepath.Join(dir, customPath+".md") {
			t.Errorf("unexpected published file: %s", entryFile)
		}

		defer func() {
			t.Log("remove the published entry")
			if _, err := blogsync("remove", entryFile); err != nil {
				t.Fatal(err)
			}
		}()
	})

}

func appendFile(path string, content string) error {
	fh, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	if _, err := fh.WriteString(content); err != nil {
		return err
	}
	return fh.Close()
}
