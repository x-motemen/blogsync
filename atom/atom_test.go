package atom

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	f, err := os.Open("../example/hatenablog.xml")
	assert.NoError(t, err)
	defer f.Close()

	feed, err := Parse(f)
	assert.NoError(t, err)

	assert.Equal(t, feed.Entries[0].Edited.UTC().String(), "2014-11-20 14:48:59 +0000 UTC")
}
