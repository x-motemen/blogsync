package atom

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Client struct {
	*http.Client
}

/*
func (c *Client) GetFeed(url string) (*Feed, error) {
}

func (c *Client) GetEntry(url string) (*Entry, error) {
}
*/

func (c *Client) PutEntry(url string, e *Entry) (*Entry, error) {
	body := new(bytes.Buffer)

	body.WriteString(xml.Header)
	err := xml.NewEncoder(body).Encode(e)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PUT", url, body)
	if err != nil {
		return nil, err
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		bytes, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("got [%s]: %q", resp.Status, string(bytes))
	}

	newEntry, err := ParseEntry(resp.Body)
	if err != nil {
		return nil, err
	}

	return newEntry, nil
}
