package atom

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Client wrapped *http.Client and some methods for accessing atom feed are added
type Client struct {
	*http.Client
}

// GetFeed gets the blog feed
func (c *Client) GetFeed(url string) (*Feed, error) {
	resp, err := c.http("GET", url, nil)
	if err != nil {
		return nil, err
	}

	return Parse(resp.Body)
}

// GetEntry gets the blog entry
func (c *Client) GetEntry(url string) (*Entry, error) {
	resp, err := c.http("GET", url, nil)
	if err != nil {
		return nil, err
	}

	return ParseEntry(resp.Body)
}

// PutEntry puts the blog entry
func (c *Client) PutEntry(url string, e *Entry) (*Entry, error) {
	body := new(bytes.Buffer)

	body.WriteString(xml.Header)
	err := xml.NewEncoder(body).Encode(e)
	if err != nil {
		return nil, err
	}

	resp, err := c.http("PUT", url, body)
	if err != nil {
		return nil, err
	}

	newEntry, err := ParseEntry(resp.Body)
	if err != nil {
		return nil, err
	}

	return newEntry, nil
}

// PostEntry posts the blog entry
func (c *Client) PostEntry(url string, e *Entry) (*Entry, error) {
	body, err := entryBody(e)
	if err != nil {
		return nil, err
	}

	resp, err := c.http("POST", url, body)
	if err != nil {
		return nil, err
	}

	newEntry, err := ParseEntry(resp.Body)
	if err != nil {
		return nil, err
	}

	return newEntry, nil
}

func entryBody(e *Entry) (*bytes.Buffer, error) {
	body := new(bytes.Buffer)

	body.WriteString(xml.Header)
	err := xml.NewEncoder(body).Encode(e)
	if err != nil {
		return nil, err
	}

	return body, nil
}

var blogsyncDebug = os.Getenv("BLOGSYNC_DEBUG") != ""

var debugLogger = sync.OnceValue(func() *slog.Logger {
	var w io.Writer = os.Stderr
	cached, err := os.UserCacheDir()
	if err == nil {
		logf := filepath.Join(cached, "blogsync", "tracedump.log")
		if err := os.MkdirAll(filepath.Dir(logf), 0755); err == nil {
			if f, err := os.OpenFile(logf, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644); err == nil {
				log.Printf("trace dumps are output to %s\n", logf)
				w = f
			}
		}
	}
	return slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
})

func (c *Client) http(method, url string, body io.Reader) (resp *http.Response, err error) {
	if blogsyncDebug {
		var reqBody, resBody string
		if body != nil {
			bb, err := io.ReadAll(body)
			if err != nil {
				return nil, err
			}
			reqBody = string(bb)
			body = strings.NewReader(reqBody)
		}
		defer func() {
			if err != nil {
				return
			}
			bb, rerr := io.ReadAll(resp.Body)
			if rerr != nil {
				err = rerr
				resp.Body.Close()
				return
			}
			resp.Body = io.NopCloser(bytes.NewReader(bb))
			resBody = string(bb)

			debugLogger().Debug("traceDump",
				slog.String("method", method),
				slog.String("url", url),
				slog.String("request", reqBody),
				slog.Int("status", resp.StatusCode),
				slog.String("response", resBody))
		}()
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	resp, err = c.Client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 300 {
		bytes, _ := io.ReadAll(resp.Body)
		return resp, fmt.Errorf("got [%s]: %q", resp.Status, string(bytes))
	}

	return resp, nil
}
