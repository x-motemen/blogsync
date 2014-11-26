package atom

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestParse(t *testing.T) {
	f, err := os.Open("../example/hatenablog.xml")
	assert.NoError(t, err)

	feed, err := Parse(f)
	assert.NoError(t, err)

	assert.Equal(t, feed.Entries[0].Edited.String(), "2014-11-20 23:48:59 +0900 JST")
}
