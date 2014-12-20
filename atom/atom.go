package atom

import (
	"encoding/xml"
	"io"
	"time"
)

type Feed struct {
	Links    Links   `xml:"link"`
	Title    string  `xml:"title"`
	Subtitle string  `xml:"subtitle"`
	Entries  []Entry `xml:"entry"`
}

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
}

type Link struct {
	Rel  string `xml:"rel,attr"`
	Href string `xml:"href,attr"`
}

type Links []Link

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

func ParseEntry(r io.Reader) (*Entry, error) {
	entry := &Entry{}
	err := xml.NewDecoder(r).Decode(entry)
	if err != nil {
		return nil, err
	}

	return entry, nil
}

func (links Links) Find(rel string) *Link {
	for _, link := range links {
		if link.Rel == rel {
			return &link
		}
	}

	return nil
}
