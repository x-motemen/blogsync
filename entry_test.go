package main

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

func TestEntryFromReader(t *testing.T) {
	f, err := os.Open("example/data/karimen.hatenablog.com/entry/2012/12/18/000000.md")
	assert.NoError(t, err)

	jst, err := time.LoadLocation("Asia/Tokyo")
	assert.NoError(t, err)

	e, err := entryFromReader(f)
	assert.NoError(t, err)

	assert.Equal(t, e.Title, "所内 #2")

	assert.True(t, e.Date.Equal(time.Date(2012, 12, 18, 0, 0, 0, 0, jst)))
}
