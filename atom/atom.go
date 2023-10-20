package atom

import (
	"encoding/xml"
	"io"
	"time"
)

// Feed represents atom feed
type Feed struct {
	Links    Links   `xml:"link"`
	Title    string  `xml:"title"`
	Subtitle string  `xml:"subtitle"`
	Entries  []Entry `xml:"entry"`
}

// Entry represents atom entry
type Entry struct {
	XMLName   xml.Name   `xml:"http://www.w3.org/2005/Atom entry"`
	ID        string     `xml:"id,omitempty"`
	Links     Links      `xml:"link"`
	Author    Author     `xml:"author,omitempty"`
	Title     string     `xml:"title"`
	Updated   *time.Time `xml:"updated,omitempty"`
	Published *time.Time `xml:"published,omitempty"`
	Edited    *time.Time `xml:"edited,omitempty"`
	Content   Content    `xml:"content"`
	Category  []Category `xml:"category,omitempty"`
	Control   *Control   `xml:"http://www.w3.org/2007/app control,omitempty"`
	CustomURL string     `xml:"http://www.hatena.ne.jp/info/xmlns#hatenablog custom-url,omitempty"`
}

// Link represents atom link
type Link struct {
	Rel  string `xml:"rel,attr"`
	Href string `xml:"href,attr"`
}

// Links represents atom links
type Links []Link

// Author represents atom author
type Author struct {
	Name string `xml:"name"`
}

// Content represents atom content
type Content struct {
	Type    string `xml:"type,attr,omitempty"`
	Content string `xml:",chardata"`
}

// Category represents atom category
type Category struct {
	Term string `xml:"term,attr"`
}

// Control represents atom control
type Control struct {
	Draft   string `xml:"http://www.w3.org/2007/app draft"`
	Preview string `xml:"http://www.w3.org/2007/app preview"`
}

// Parse parses an atom xml from r and returns Feed
func Parse(r io.Reader) (*Feed, error) {
	feed := &Feed{}
	err := xml.NewDecoder(r).Decode(feed)
	if err != nil {
		return nil, err
	}

	return feed, nil
}

// ParseEntry parses an atom xml from r and returns Entry
func ParseEntry(r io.Reader) (*Entry, error) {
	entry := &Entry{}
	err := xml.NewDecoder(r).Decode(entry)
	if err != nil {
		return nil, err
	}

	return entry, nil
}

// Find finds the Link from links by a specified string argument
func (links Links) Find(rel string) *Link {
	for _, link := range links {
		if link.Rel == rel {
			return &link
		}
	}

	return nil
}
