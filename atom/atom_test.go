package atom

import (
	"os"
	"testing"
)

func TestParse(t *testing.T) {
	f, err := os.Open("../example/hatenablog.xml")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	feed, err := Parse(f)
	if err != nil {
		t.Fatal(err)
	}

	expect := "2014-11-20 14:48:59 +0000 UTC"
	if g := feed.Entries[0].Edited.UTC().String(); g != expect {
		t.Errorf("expect: %s, but got: %s", expect, g)
	}
}
