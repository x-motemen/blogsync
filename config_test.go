package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetConfig(t *testing.T) {
	config := &Config{
		Default: &BlogConfig{
			LocalRoot: "./data",
		},
		Blogs: map[string]*BlogConfig{
			"blog.example.com": {
				RemoteRoot: "blog.example.com",
				Username:   "xxx",
				Password:   "yyy",
			},
		},
	}

	c := config.Get("blog.example.com")
	assert.NotNil(t, c)
	assert.Equal(t, c.LocalRoot, "./data")
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
	conf, err := LoadConfig(r)
	assert.Nil(t, err)
	assert.Equal(t, conf.Default.LocalRoot, "./data")
	assert.Equal(t, conf.Blogs["blog1.example.com"].Username, "blog1")
	assert.Equal(t, conf.Blogs["blog1.example.com"].RemoteRoot, "blog1.example.com")
}
