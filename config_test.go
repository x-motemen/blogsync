package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetConfig(t *testing.T) {
	c := &config{
		Default: &blogConfig{
			LocalRoot: "./data",
		},
		Blogs: map[string]*blogConfig{
			"blog.example.com": {
				RemoteRoot: "blog.example.com",
				Username:   "xxx",
				Password:   "yyy",
			},
		},
	}

	bc := c.Get("blog.example.com")
	assert.NotNil(t, bc)
	assert.Equal(t, bc.LocalRoot, "./data")
}

func TestLoadConfig(t *testing.T) {
	r := bytes.NewReader([]byte(
		`---
default:
  local_root: ./data
blog1.example.com:
  username: blog1
blog2.example.com:
  local_root: ./blog2`,
	))
	c, err := loadConfig(r)
	assert.Nil(t, err)
	assert.Equal(t, c.Default.LocalRoot, "./data")
	assert.Equal(t, c.Blogs["blog1.example.com"].Username, "blog1")
	assert.Equal(t, c.Blogs["blog1.example.com"].RemoteRoot, "blog1.example.com")
}
