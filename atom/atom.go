package atom

import (
	"encoding/xml"
	"io"
	"time"
)

type Feed struct {
	Links    []Link  `xml:"link"`
	Title    string  `xml:"title"`
	Subtitle string  `xml:"subtitle"`
	Entries  []Entry `xml:"entry"`
}

type Entry struct {
	XMLName   xml.Name  `xml:"entry"`
	ID        string    `xml:"id,omitempty"`
	Links     []Link    `xml:"link"`
	Author    Author    `xml:"author,omitempty"`
	Title     string    `xml:"title"`
	Updated   time.Time `xml:"updated"`
	Published time.Time `xml:"published"`
	Content   Content   `xml:"content"`
	XMLNs     string    `xml:"xmlns,attr"`
}

type Link struct {
	Rel  string `xml:"rel,attr"`
	Href string `xml:"href,attr"`
}

type Author struct {
	Name string `xml:"name"`
}

type Content struct {
	Type    string `xml:"type,attr,omitempty"`
	Content string `xml:",chardata"`
}

func Parse(r io.Reader) (*Feed, error) {
	feed := &Feed{}
	err := xml.NewDecoder(r).Decode(feed)
	if err != nil {
		return nil, err
	}

	return feed, nil
}

// utility
func FindLink(rel string, links []Link) *Link {
	for _, link := range links {
		if link.Rel == rel {
			return &link
		}
	}

	return nil
}
